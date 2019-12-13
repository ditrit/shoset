package main

import (
	"crypto/tls"
	"fmt"
	"time"
)

var (
	tlsConfigClient = tls.Config{InsecureSkipVerify: true}
)

// GSClient : first test on event client socket
type GSClient struct {
	socket    *tls.Conn
	reconnect chan bool
	sndEvt    chan string
	address   string
}

// NewGSClient : constructor
func NewGSClient(address string) *GSClient {
	s := new(GSClient)
	s.socket = new(tls.Conn)
	s.sndEvt = make(chan string)
	s.reconnect = make(chan bool)
	s.address = address
	fmt.Print("GSClient initialized\n")
	return s
}

// Run : handler for the socket
func (s *GSClient) Run() {
	fmt.Print("Run launched\n")
	n := 0

	for {
		n++
		fmt.Printf("new step %d", n)
		conn, err := tls.Dial("tcp", s.address, &tlsConfigClient)
		if err != nil {
			fmt.Println("Failed to connect:", err.Error())
			fmt.Printf("Trying reset the connection (%d)...\n", n)
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			fmt.Printf("connection succeeded")
			s.socket = conn

			// receive form Socket goroutine
			go func() {
				buffer := make([]byte, 1024)
				for {
					bytesRead, err := s.socket.Read(buffer)
					if err != nil {
						fmt.Println("Failed to read:", err.Error())
						fmt.Println("Trying reset the connection...")
						s.reconnect <- true
						return
					} else {
						fmt.Printf("Received event form sovket : %s", string(bytesRead))
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

func main() {
	test := NewGSClient("localhost:8100")
	test.Run()
}
