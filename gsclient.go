package main

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
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
	done  chan bool
	m     sync.RWMutex
}

// NewGSClient : constructor
func NewGSClient(address string) (*GSClient, error) {
	s := new(GSClient)
	s.conns = make(map[string]*GSClientConn)
	s.Add(address)
	return s, nil
}

//Add : Add a new connection to a server
func (s *GSClient) Add(address string) {
	conn, _ := NewGSClientConn(address)
	s.m.Lock()
	s.conns[address] = conn
	s.m.Unlock()
}

// GSClientConn : client connection
type GSClientConn struct {
	socket    *tls.Conn
	rw        *bufio.ReadWriter
	m         sync.RWMutex
	reconnect chan bool
	sndEvt    chan msg.Event
	sndCmd    chan msg.Command
	sndRep    chan msg.Reply
	sndCfg    chan string
	address   string
}

// NewGSClientConn : constructor
func NewGSClientConn(address string) (*GSClientConn, error) {
	s := new(GSClientConn)
	s.socket = new(tls.Conn)
	s.rw = new(bufio.ReadWriter)
	s.sndEvt = make(chan msg.Event)
	s.sndCmd = make(chan msg.Command)
	s.sndRep = make(chan msg.Reply)
	s.sndCfg = make(chan string)
	s.reconnect = make(chan bool)
	s.address = address
	go s.Run()
	return s, nil
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
					switch msgType {
					case "evt":
						var event msg.Event
						s.m.Lock()
						dec := gob.NewDecoder(s.rw)
						err := dec.Decode(&event)
						s.m.Unlock()
						if err != nil {
							fmt.Println("Error decoding Event data ", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Event : \n%#v\n", event)
					case "cmd":
						var command msg.Command
						s.m.Lock()
						dec := gob.NewDecoder(s.rw)
						err := dec.Decode(&command)
						s.m.Unlock()
						if err != nil {
							fmt.Println("Error decoding Command data ", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Event : \n%#v\n", command)
					case "rep":
						var reply msg.Reply
						s.m.Lock()
						dec := gob.NewDecoder(s.rw)
						err := dec.Decode(&reply)
						s.m.Unlock()
						if err != nil {
							fmt.Println("Error decoding Reply data ", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Event : \n%#v\n", reply)
					case "cfg":
						s.m.Lock()
						config, err := s.rw.ReadString('\n')
						s.m.Unlock()
						if err != nil {
							fmt.Println("Error reading config string", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Config: \n\t%s\n", config)
					default:
						fmt.Printf("%s msg type is not implemented\n", msgType)
						s.reconnect <- true
						return
					}
				}
			}()

			// manage events
			doSelect := true
			for doSelect {
				select {
				case evt := <-s.sndEvt:
					fmt.Printf("event on chanel")
					s.m.Lock()
					s.rw.WriteString("evt\n")
					enc := gob.NewEncoder(s.rw)
					err := enc.Encode(evt)
					if err != nil {
						fmt.Println("Failed to write event:", err.Error())
						fmt.Println("Trying reset the connection...")
						break
					}
					err = s.rw.Flush()
					if err != nil {
						fmt.Printf("failed to flush")
						break
					}
					s.m.Unlock()
					fmt.Printf("sent")
				case cmd := <-s.sndCmd:
					fmt.Printf("command on chanel")
					s.m.Lock()
					s.rw.WriteString("cmd\n")
					enc := gob.NewEncoder(s.rw)
					err := enc.Encode(cmd)
					if err != nil {
						fmt.Println("Failed to write command:", err.Error())
						fmt.Println("Trying reset the connection...")
						break
					}
					err = s.rw.Flush()
					if err != nil {
						fmt.Printf("failed to flush")
						break
					}
					s.m.Unlock()
					fmt.Printf("sent")
				case rep := <-s.sndRep:
					fmt.Printf("reply on chanel")
					s.m.Lock()
					s.rw.WriteString("rep\n")
					enc := gob.NewEncoder(s.rw)
					err := enc.Encode(rep)
					if err != nil {
						fmt.Println("Failed to write reply:", err.Error())
						fmt.Println("Trying reset the connection...")
						break
					}
					err = s.rw.Flush()
					if err != nil {
						fmt.Printf("failed to flush")
						break
					}
					s.m.Unlock()
					fmt.Printf("sent")
				case cfg := <-s.sndCfg:
					fmt.Printf("config on chanel %s", cfg)
					s.m.Lock()
					_, err := s.rw.WriteString("cfg\n")
					if err != nil {
						fmt.Printf("failed to send message type 'cfg' message")
						break
					}
					_, err = s.rw.WriteString(cfg + "\n")
					if err != nil {
						fmt.Printf("failed to send message type 'cfg' message")
						break
					}
					err = s.rw.Flush()
					if err != nil {
						fmt.Printf("failed to flush")
						break
					}
					s.m.Unlock()
					fmt.Printf("sent")

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
func (s *GSClient) SendConfig(cfg string) {
	fmt.Printf("Sending a configuration : %s", cfg)
	for _, conn := range s.conns {
		conn.sndCfg <- cfg
	}
}

func client(address string) {
	c, _ := NewGSClient(address)
	go func() {
		event := msg.NewEvent("token", "bus", "started", "ok")
		c.SendEvent(event)
		command := msg.NewCommand("token", "orchestrator", "deploy", "{\"appli\": \"toto\"}")
		c.SendCommand(command)
	}()
	<-c.done
}
