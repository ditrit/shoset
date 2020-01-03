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
	rb         *msg.Reader
	wb         *msg.Writer
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
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
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

func (s *GSServerConn) receiveMsg() error {
	// read message type
	msgType, err := s.rb.ReadString()
	fmt.Printf("Type message read : %s\n", msgType)
	switch {
	case err == io.EOF:
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		s.stop <- true
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		s.stop <- true
		return errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")

	fmt.Printf("msgType '%s'", msgType)
	// read message data
	switch msgType {
	case "evt":
		var evt msg.Event
		err = s.rb.ReadEvent(&evt)
		if err != nil {
			fmt.Printf("Error!!!")
		}
		fmt.Printf("Receive %s: \n%#v.\n", msgType, evt)
		//s.gsClient.msgQueue.PushEvent(*evt)
	case "cmd":
		var cmd msg.Command
		err = s.rb.ReadCommand(&cmd)
		fmt.Printf("Receive %s: \n%#v\n", msgType, cmd)
		//s.gsClient.msgQueue.PushCommand(*cmd)
	case "rep":
		var rep msg.Reply
		err = s.rb.ReadReply(&rep)
		fmt.Printf("Receive %s: \n%#v\n", msgType, rep)
		//s.gsClient.msgQueue.PushReply(*rep)
	case "cfg":
		var cfg msg.Config
		err = s.rb.ReadConfig(&cfg)
		fmt.Printf("Receive %s: \n%#v\n", msgType, cfg)
		//s.gsClient.msgQueue.PushConfig(*cfg)
	default:
		fmt.Printf("%s msg type is not implemented\n", msgType)
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		return errors.New("non implemented message " + msgType)
	}
	if err != nil {
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		s.stop <- true
		return errors.New("receiveMsg : unable to decode a message of type  " + msgType)
	}
	return err
}

// Run : handler for the connection
func (s *GSServerConn) Run() {
	s.rb = msg.NewReader(s.socket)
	s.wb = msg.NewWriter(s.socket)

	// receive messages
	for {
		err := s.receiveMsg()
		if err != nil {
			return
		}
	}

}

// SendEvent : send an event...
// event is sent on each connection
func (s *GSServer) SendEvent(evt *msg.Event) {
	fmt.Print("Sending event.\n")
	for _, conn := range s.conns {
		conn.wb.WriteString("evt")
		conn.wb.WriteEvent(*evt)
	}
}

// SendCommand : Send a message
// todo : manage routing
//    identify relevant targets (routing info matches identity)
//    then try on each instance until success
func (s *GSServer) SendCommand(cmd *msg.Command) {
	fmt.Print("Sending command.\n")

	for _, conn := range s.conns {
		conn.wb.WriteString("cmd")
		conn.wb.WriteCommand(*cmd)
	}
}

// SendReply :
func (s *GSServer) SendReply(rep *msg.Reply) {
	fmt.Print("Sending reply.\n")

	for _, conn := range s.conns {
		conn.wb.WriteString("rep")
		conn.wb.WriteReply(*rep)
	}
}

// SendConfig :
func (s *GSServer) SendConfig() {
	fmt.Print("Sending configuration.\n")

	cfg := msg.NewConfig(s.name)
	for _, conn := range s.conns {
		conn.wb.WriteString("evt")
		conn.wb.WriteConfig(*cfg)
	}
}

func server(name string, address string) {
	s, err := NewGSServer(name, address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	}
	go func() {
		time.Sleep(time.Second * time.Duration(5))
		event := msg.NewEvent("bus", "starting", "ok")
		s.SendEvent(event)
		time.Sleep(time.Millisecond * time.Duration(200))
		event = msg.NewEvent("bus", "started", "ok")
		s.SendEvent(event)
		command := msg.NewCommand("bus", "register", "{\"topic\": \"toto\"}")
		s.SendCommand(command)
		reply := msg.NewReply(command, "success", "OK")
		s.SendReply(reply)

	}()
	<-s.done
}
