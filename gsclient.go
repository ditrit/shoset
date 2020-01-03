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
	qEvents   *msg.Queue
	qCommands *msg.Queue
	qReplies  *msg.Queue
	qConfigs  *msg.Queue
}

// NewGSClient : constructor
func NewGSClient(name string, address string) (*GSClient, error) {
	s := new(GSClient)
	s.name = name
	s.conns = make(map[string]*GSClientConn)
	s.qEvents = msg.NewQueue()
	s.qCommands = msg.NewQueue()
	s.qReplies = msg.NewQueue()
	s.qConfigs = msg.NewQueue()
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

// WaitEvent :
func (s *GSClient) WaitEvent(events *msg.Iterator, topicName string, eventName string, timeout int) *msg.Event {
	term := make(chan *msg.Event, 1)
	cont := true
	go func() {
		for cont {
			message := events.Get()
			if message != nil {
				event := (*message).(msg.Event)
				if topicName == event.GetTopic() && eventName == event.GetEvent() {
					term <- &event
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

// WaitCommand : uniquement au sein d'un connecteur a priori
func (s *GSClient) WaitCommand(commands *msg.Iterator, commandName string, timeout int) *msg.Command {
	term := make(chan *msg.Command, 1)
	cont := true
	go func() {
		for cont {
			message := commands.Get()
			if message != nil {
				command := (*message).(msg.Command)
				if commandName == command.GetCommand() {
					term <- &command
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

// WaitReply :
func (s *GSClient) WaitReply(replies *msg.Iterator, commandUUID string, timeout int) *msg.Reply {
	term := make(chan *msg.Reply, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				reply := (*message).(msg.Reply)
				if commandUUID == reply.GetCmdUUID() {
					term <- &reply
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

// WaitConfig :
func (s *GSClient) WaitConfig(replies *msg.Iterator, commandUUID string, timeout int) *msg.Config {
	term := make(chan *msg.Config, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				config := (*message).(msg.Config)
				term <- &config
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

func client(name string, address string) {
	c, _ := NewGSClient(name, address)
	time.Sleep(time.Second * time.Duration(1))
	go func() {
		command := msg.NewCommand("orchestrator", "deploy", "{\"appli\": \"toto\"}")
		c.SendCommand(command)
		event := msg.NewEvent("bus", "coucou", "ok")
		c.SendEvent(event)

		events := msg.NewIterator(c.qEvents)
		defer events.Close()
		rec := c.WaitEvent(events, "bus", "started", 20)
		if rec != nil {
			fmt.Printf(">Received Event: \n%#v\n", *rec)
		} else {
			fmt.Print("Timeout expired !")
		}
		events2 := msg.NewIterator(c.qEvents)
		defer events.Close()
		rec2 := c.WaitEvent(events2, "bus", "starting", 20)
		if rec2 != nil {
			fmt.Printf(">Received Event 2: \n%#v\n", *rec2)
		} else {
			fmt.Print("Timeout expired  2 !")
		}
	}()

	<-c.done
}
