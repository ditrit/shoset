package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"./msg"
)

var certPath = "./cert.pem"
var keyPath = "./key.pem"

// GSServer : first test on event client socket
type GSServer struct {
	conns   map[string]*GSServerConn
	name    string
	done    chan bool
	m       sync.RWMutex
	config  *tls.Config
	address string
}

// NewGSServer : constructor
func NewGSServer(name string, address string) (*GSServer, error) {
	s := new(GSServer)
	s.name = name
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		fmt.Printf("tls init failed %s\n", err)
		return nil, err
	}
	s.config = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	s.address = address
	s.conns = make(map[string]*GSServerConn)
	go s.Run()
	return s, nil
}

// GSServerConn : server connection
type GSServerConn struct {
	socket     *tls.Conn
	localName  string
	remoteName string
	address    string
	gsServer   *GSServer
	rb         *SafeReader
	wb         *SafeWriter
	sndEvt     chan msg.Event
	sndCmd     chan msg.Command
	sndRep     chan msg.Reply
	sndCfg     chan msg.Config
	stop       chan bool
}

//add : Add a new connection from a client
func (s *GSServer) add(tlsConn *tls.Conn) (*GSServerConn, error) {
	conn := new(GSServerConn)
	conn.socket = tlsConn
	conn.gsServer = s
	conn.address = tlsConn.RemoteAddr().String()
	conn.localName = s.name
	s.m.Lock()
	s.conns[conn.address] = conn
	s.m.Unlock()
	conn.rb = new(SafeReader)
	conn.wb = new(SafeWriter)
	conn.sndEvt = make(chan msg.Event)
	conn.sndCmd = make(chan msg.Command)
	conn.sndRep = make(chan msg.Reply)
	conn.sndCfg = make(chan msg.Config)
	conn.stop = make(chan bool)

	//go conn.Run()
	return conn, nil
}

// Run : handler for the socket
func (s *GSServer) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Println("Failed to bind:", err.Error())
		fmt.Print("GSServer initialized\n")
		return err
	}
	defer listener.Close()

	for {
		unencConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("server: accept %s", err)
			break
		}
		tlsConn := tls.Server(unencConn, s.config)
		conn, _ := s.add(tlsConn)
		fmt.Printf("GSServer : accepted from %s", conn.address)
		go conn.Run()
	}
	return nil
}

func (s *GSServerConn) receiveMsg() (interface{}, error) {
	var data interface{}

	// read message type
	msgType, err := s.rb.ReadString()
	switch {
	case err == io.EOF:
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		s.stop <- true
		return nil, errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		s.stop <- true
		return nil, errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")

	// read message data
	switch msgType {
	case "evt":
		data = new(msg.Event)
	case "cmd":
		data = new(msg.Command)
	case "rep":
		data = new(msg.Reply)
	case "cfg":
		data = new(msg.Config)
	default:
		fmt.Printf("%s msg type is not implemented\n", msgType)
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		return nil, errors.New("non implemented message " + msgType)
	}
	err = s.rb.ReadMessage(&data)
	if err != nil {
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		s.stop <- true
		return nil, errors.New("receiveMsg : unable to decode a message of type  " + msgType)
	}
	fmt.Printf("Receive %s: \n%#v\n", msgType, data)
	return &data, err
}

func (s *GSServerConn) sendMsg() error {
	var data interface{}
	var msgType string
	select {
	case evt := <-s.sndEvt:
		data = &evt
		msgType = "evt"
	case cmd := <-s.sndCmd:
		data = &cmd
		msgType = "cmd"
	case rep := <-s.sndRep:
		data = &rep
		msgType = "rep"
	case cfg := <-s.sndCfg:
		data = &cfg
		msgType = "cfg"
	case <-s.stop:
		return errors.New("sendMsg : close connection  ")
	}

	_, err := s.wb.WriteString(msgType)
	if err != nil {
		fmt.Printf("sendMsg : can not send the message type  \n")
		return errors.New("sendMsg : can not send the message type ")
	}

	err = s.wb.WriteMessage(&data)
	if err != nil {
		fmt.Printf("sendMsg : can not send the message  \n")
		return errors.New("sendMsg : can not send data part of the message")
	}
	err = s.wb.Flush()
	if err != nil {
		fmt.Printf("sendMsg : can not flush  \n")
		return errors.New("sendMsg : can not flush")
	}
	return nil
}

// Run : handler for the connection
func (s *GSServerConn) Run() {
	s.rb = NewSafeReader(s.socket)
	s.wb = NewSafeWriter(s.socket)

	// receive messages
	go func() {
		for {
			_, err := s.receiveMsg()
			if err != nil {
				return
			}
		}
	}()

	// send messages
	doSelect := true
	for doSelect {
		err := s.sendMsg()
		if err != nil {
			fmt.Printf("error sending message : %s", err)
			doSelect = false
		}
	}
}

// SendEvent : send an event...
// event is sent on each connection
func (s *GSServer) SendEvent(evt *msg.Event) {
	for _, conn := range s.conns {
		conn.sndEvt <- *evt
	}
}

// SendCommand : Send a message
// todo : manage routing
//    identify relevant targets (routing info matches identity)
//    then try on each instance until success
func (s *GSServer) SendCommand(cmd *msg.Command) {
	for _, conn := range s.conns {
		conn.sndCmd <- *cmd
	}
}

// SendReply :
func (s *GSServer) SendReply(rep *msg.Reply) {
	for _, conn := range s.conns {
		conn.sndRep <- *rep
	}
}

// SendConfig :
func (s *GSServer) SendConfig() {
	cfg := msg.NewConfig(s.name)
	for _, conn := range s.conns {
		conn.sndCfg <- *cfg
	}
}

func server(name string, address string) {
	s, err := NewGSServer(name, address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	}
	go func() {
		time.Sleep(time.Second * time.Duration(5))
		event := msg.NewEvent("token", "bus", "started", "ok")
		s.SendEvent(event)
		command := msg.NewCommand("token", "bus", "register", "{\"topic\": \"toto\"}")
		s.SendCommand(command)
		reply := command.NewReply("success", "OK")
		s.SendReply(reply)
	}()
	<-s.done
}
