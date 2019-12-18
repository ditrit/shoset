package main

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
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
	gsClient   *GSClient
	rw         *bufio.ReadWriter
	m          sync.RWMutex
	reconnect  chan bool
	sndEvt     chan msg.Event
	sndCmd     chan msg.Command
	sndRep     chan msg.Reply
	sndCfg     chan msg.Config
	address    string
}

//Add : Add a new connection to a server
func (s *GSClient) Add(address string) (*GSClientConn, error) {
	conn := new(GSClientConn)
	conn.gsClient = s
	s.m.Lock()
	s.conns[address] = conn
	s.m.Unlock()
	conn.socket = new(tls.Conn)
	conn.rw = new(bufio.ReadWriter)
	conn.sndEvt = make(chan msg.Event)
	conn.sndCmd = make(chan msg.Command)
	conn.sndRep = make(chan msg.Reply)
	conn.sndCfg = make(chan msg.Config)
	conn.reconnect = make(chan bool)
	conn.address = address
	conn.localName = s.name
	go conn.Run()
	return conn, nil
}

func (s *GSClientConn) getMsg(data interface{}) error {
	s.m.Lock()
	dec := gob.NewDecoder(s.rw)
	err := dec.Decode(&data)
	s.m.Unlock()
	if err != nil {
		fmt.Println("Error decoding received message ", err)
		s.reconnect <- true
	}
	return err
}

func (s *GSClientConn) receiveMsg(msgType string) (interface{}, error) {
	var data interface{}
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
		return nil, errors.New("non implemented message " + msgType)
	}
	s.m.Lock()
	dec := gob.NewDecoder(s.rw)
	err := dec.Decode(&data)
	s.m.Unlock()
	if err != nil {
		fmt.Println("Error decoding received message ", err)
		s.reconnect <- true
	}
	fmt.Printf("Receive %s: \n%#v\n", msgType, data)
	return &data, err
}

func (s *GSClientConn) sendMsg(typeMsg string, data interface{}) error {
	fmt.Printf("send message")
	s.m.Lock()
	s.rw.WriteString(typeMsg + "\n")
	enc := gob.NewEncoder(s.rw)
	err := enc.Encode(&data)
	if err != nil {
		fmt.Println("Failed to write message:", err.Error())
		fmt.Println("Trying reset the connection...")
		return err
	}
	err = s.rw.Flush()
	if err != nil {
		fmt.Printf("failed to flush")
		return err
	}
	s.m.Unlock()
	fmt.Printf("sent")
	return nil
}

// Run : handler for the socket
func (s *GSClientConn) Run() {
	fmt.Print("Run launched\n")
	n := 0

	for {
		conn, err := tls.Dial("tcp", s.address, &tlsConfigClient)
		if err != nil {
			fmt.Println("Failed to connect:", err.Error())
			fmt.Printf("Trying reset the connection (%d)...\n", n)
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			s.socket = conn
			s.rw = bufio.NewReadWriter(bufio.NewReader(s.socket), bufio.NewWriter(s.socket))

			// receive from Socket goroutine
			go func() {
				for {
					fmt.Print("Receive command ")
					s.m.Lock()
					msgType, err := s.rw.ReadString('\n')
					s.m.Unlock()
					switch {
					case err == io.EOF:
						fmt.Println("Reached EOF - close this connection.\n ---")
						s.reconnect <- true
						return
					case err != nil:
						fmt.Println("Failed to read:", err.Error())
						s.reconnect <- true
						return
					}
					msgType = strings.Trim(msgType, "\n")
					fmt.Println("'" + msgType + "'")

					s.receiveMsg(msgType)
				}
			}()

			// manage events
			doSelect := true
			for doSelect {
				select {
				case evt := <-s.sndEvt:
					if s.sendMsg("evt", evt) != nil {
						break
					}
				case cmd := <-s.sndCmd:
					if s.sendMsg("cmd", cmd) != nil {
						break
					}
				case rep := <-s.sndRep:
					if s.sendMsg("rep", rep) != nil {
						break
					}
				case cfg := <-s.sndCfg:
					if s.sendMsg("cfg", cfg) != nil {
						break
					}
				case <-s.reconnect:
					fmt.Print("reconnect received")
					doSelect = false
					break
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
