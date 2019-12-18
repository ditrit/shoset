package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"./msg"
)

var (
	tlsConfigClient = tls.Config{InsecureSkipVerify: true}
)

//GSClient : client gandalf Socket
type GSClient struct {
	conns map[string]*GSClientConn
	name  string
	done  chan bool
	m     sync.RWMutex
}

// NewGSClient : constructor
func NewGSClient(name string, address string) (*GSClient, error) {
	s := new(GSClient)
	s.name = name
	s.conns = make(map[string]*GSClientConn)
	_, e := s.Add(address)
	return s, e
}

// GSClientConn : client connection
type GSClientConn struct {
	socket     *tls.Conn
	localName  string
	remoteName string
	address    string
	gsClient   *GSClient
	rb         *SafeReader
	wb         *SafeWriter
	sndEvt     chan msg.Event
	sndCmd     chan msg.Command
	sndRep     chan msg.Reply
	sndCfg     chan msg.Config
	stop       chan bool
}

//Add : Add a new connection to a server
func (s *GSClient) Add(address string) (*GSClientConn, error) {
	conn := new(GSClientConn)
	conn.gsClient = s
	s.m.Lock()
	s.conns[address] = conn
	s.m.Unlock()
	conn.socket = new(tls.Conn)
	conn.rb = new(SafeReader)
	conn.wb = new(SafeWriter)
	conn.sndEvt = make(chan msg.Event)
	conn.sndCmd = make(chan msg.Command)
	conn.sndRep = make(chan msg.Reply)
	conn.sndCfg = make(chan msg.Config)
	conn.stop = make(chan bool)
	conn.address = address
	conn.localName = s.name
	go conn.Run()
	return conn, nil
}

func (s *GSClientConn) receiveMsg() (interface{}, error) {
	var data interface{}

	// read message type
	msgType, err := s.rb.ReadString()
	switch {
	case err == io.EOF:
		s.stop <- true
		return nil, errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
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
		return nil, errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	err = s.rb.ReadMessage(&data)
	if err != nil {
		s.stop <- true
		return nil, errors.New("receiveMsg : unable to decode a message of type  " + msgType)
	}
	fmt.Printf("Receive %s: \n%#v\n", msgType, data)
	return &data, err
}

func (s *GSClientConn) sendMsg() error {
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
		return errors.New("sendMsg : reconnect expected  ")
	}

	_, err := s.wb.WriteString(msgType)
	if err != nil {
		fmt.Printf("sendMsg : can not send the message type  \n")
		return errors.New("sendMsg : can not send the message type ")
	}
	err = s.wb.WriteMessage(&data)
	if err != nil {
		return errors.New("sendMsg : can not send the message  ")
	}
	err = s.wb.Flush()
	if err != nil {
		return errors.New("sendMsg : can not flush")
	}
	return nil
}

// Run : handler for the socket
func (s *GSClientConn) Run() {
	for {
		conn, err := tls.Dial("tcp", s.address, &tlsConfigClient)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			s.socket = conn
			s.rb = NewSafeReader(s.socket)
			s.wb = NewSafeWriter(s.socket)

			// receive messages
			go func() {
				for {
					s.receiveMsg()
				}
			}()

			// send messages
			doSelect := true
			for doSelect {
				err = s.sendMsg()
				if err != nil {
					doSelect = false
				}
			}
		}
	}
}

// SendEvent : send an event...
// event is sent on each connection
func (s *GSClient) SendEvent(evt *msg.Event) {
	for _, conn := range s.conns {
		conn.sndEvt <- *evt
	}
}

// SendCommand : Send a message
// todo : manage routing
//    identify relevant targets (routing info matches identity)
//    then try on each instance until success
func (s *GSClient) SendCommand(cmd *msg.Command) {
	for _, conn := range s.conns {
		conn.sndCmd <- *cmd
	}
}

// SendReply :
func (s *GSClient) SendReply(rep *msg.Reply) {
	for _, conn := range s.conns {
		conn.sndRep <- *rep
	}
}

// SendConfig :
func (s *GSClient) SendConfig() {
	fmt.Print("Sending configuration.\n")

	cfg := msg.NewConfig(s.name)
	for _, conn := range s.conns {
		conn.sndCfg <- *cfg
	}
}

func client(name string, address string) {
	c, _ := NewGSClient(name, address)
	go func() {
		event := msg.NewEvent("token", "bus", "started", "ok")
		c.SendEvent(event)
		command := msg.NewCommand("token", "orchestrator", "deploy", "{\"appli\": \"toto\"}")
		c.SendCommand(command)
	}()
	<-c.done
}
