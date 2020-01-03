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
	conns     map[string]*GSClientConn
	name      string
	done      chan bool
	m         sync.RWMutex
	qEvents   msg.Queue
	qCommands msg.Queue
	qReplies  msg.Queue
	qConfigs  msg.Queue
}

// NewGSClient : constructor
func NewGSClient(name string, address string) (*GSClient, error) {
	s := new(GSClient)
	s.name = name
	s.conns = make(map[string]*GSClientConn)
	s.qEvents.Init()
	s.qCommands.Init()
	s.qReplies.Init()
	s.qConfigs.Init()
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
	rb         *msg.Reader
	wb         *msg.Writer

	stop chan bool
}

//Add : Add a new connection to a server
func (s *GSClient) Add(address string) (*GSClientConn, error) {
	conn := new(GSClientConn)
	conn.gsClient = s
	s.m.Lock()
	s.conns[address] = conn
	s.m.Unlock()
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	conn.stop = make(chan bool)
	conn.address = address
	conn.localName = s.name
	go conn.Run()
	return conn, nil
}

func (s *GSClientConn) receiveMsg() error {
	// read message type
	msgType, err := s.rb.ReadString()
	switch {
	case err == io.EOF:
		s.stop <- true
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		s.stop <- true
		return errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")

	// read message data
	fmt.Printf("Read message and push if into buffer")
	switch msgType {
	case "evt":
		var evt msg.Event
		err = s.rb.ReadEvent(&evt)
		s.gsClient.qEvents.Push(evt)
	case "cmd":
		var cmd msg.Command
		err = s.rb.ReadCommand(&cmd)
		s.gsClient.qCommands.Push(cmd)
	case "rep":
		var rep msg.Reply
		err = s.rb.ReadReply(&rep)
		s.gsClient.qReplies.Push(rep)
	case "cfg":
		var cfg msg.Config
		err = s.rb.ReadConfig(&cfg)
		s.gsClient.qConfigs.Push(cfg)
	default:
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	if err != nil {
		s.stop <- true
		return errors.New("receiveMsg : unable to decode a message of type  " + msgType)
	}
	return err
}

// Run : handler for the socket
func (s *GSClientConn) Run() {
	for {
		conn, err := tls.Dial("tcp", s.address, &tlsConfigClient)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			s.socket = conn
			s.rb = msg.NewReader(s.socket)
			s.wb = msg.NewWriter(s.socket)

			// receive messages
			for {
				fmt.Printf("Receive Msg\n")
				s.receiveMsg()
			}
		}
	}
}

// SendEvent : send an event...
// event is sent on each connection
func (s *GSClient) SendEvent(evt *msg.Event) {
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
func (s *GSClient) SendCommand(cmd *msg.Command) {
	fmt.Print("Sending command.\n")

	for _, conn := range s.conns {
		conn.wb.WriteString("cmd")
		conn.wb.WriteCommand(*cmd)
	}
}

// SendReply :
func (s *GSClient) SendReply(rep *msg.Reply) {
	fmt.Print("Sending reply.\n")

	for _, conn := range s.conns {
		conn.wb.WriteString("rep")
		conn.wb.WriteReply(*rep)
	}
}

// SendConfig :
func (s *GSClient) SendConfig() {
	fmt.Print("Sending configuration.\n")

	cfg := msg.NewConfig(s.name)
	for _, conn := range s.conns {
		conn.wb.WriteString("evt")
		conn.wb.WriteConfig(*cfg)
	}
}

func client(name string, address string) {
	c, _ := NewGSClient(name, address)
	time.Sleep(time.Second * time.Duration(1))
	go func() {
		command := msg.NewCommand("orchestrator", "deploy", "{\"appli\": \"toto\"}")
		c.SendCommand(command)
		event := msg.NewEvent("bus", "started", "ok")
		c.SendEvent(event)
	}()

	<-c.done
}

// WaitEvent :
func (s *GSClient) WaitEvent() {

}
