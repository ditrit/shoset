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
	reconnect chan bool
	sndEvt    chan string
	sndCmd    chan string
	address   string
}

// NewGSClientConn : constructor
func NewGSClientConn(address string) (*GSClientConn, error) {
	s := new(GSClientConn)
	s.socket = new(tls.Conn)
	s.sndEvt = make(chan string)
	s.sndCmd = make(chan string)
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

			// receive from Socket goroutine
			go func() {
				rw := bufio.NewReadWriter(bufio.NewReader(s.socket), bufio.NewWriter(s.socket))
				for {
					fmt.Print("Receive command ")
					msgType, err := rw.ReadString('\n')
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
						dec := gob.NewDecoder(rw)
						err := dec.Decode(&event)
						if err != nil {
							fmt.Println("Error decoding Event data ", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Event : \n%#v\n", event)
					case "cmd":
						var command msg.Command
						dec := gob.NewDecoder(rw)
						err := dec.Decode(&command)
						if err != nil {
							fmt.Println("Error decoding Command data ", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Event : \n%#v\n", command)
					case "rep":
						var reply msg.Reply
						dec := gob.NewDecoder(rw)
						err := dec.Decode(&reply)
						if err != nil {
							fmt.Println("Error decoding Reply data ", err)
							s.reconnect <- true
							return
						}
						fmt.Printf("Event : \n%#v\n", reply)
					case "cfg":
						config, err := rw.ReadString('\n')
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
				case event := <-s.sndEvt:
					_, err := s.socket.Write([]byte(event))
					if err != nil {
						fmt.Println("Failed to write:", err.Error())
						fmt.Println("Trying reset the connection...")
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

// Send : Send a message
func (s *GSClientConn) Send(cmd string) {
	msgCmd := msg.NewCommand("token", "connType", cmd, "{payload string}")
	fmt.Printf(msgCmd.Uuid)
}

func client(address string) {
	c, _ := NewGSClient(address)
	go func() {
		msg := "un petit test"
		for _, conn := range c.conns {
			conn.Send(msg)
		}
	}()
	<-c.done
}
