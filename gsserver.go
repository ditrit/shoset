package main

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"./msg"
)

var certPath = "./cert.pem"
var keyPath = "./key.pem"

// GSServer : first test on event client socket
type GSServer struct {
	config  *tls.Config
	address string
	conns   map[string]*tls.Conn

	m sync.RWMutex
}

// NewGSServer : constructor
func NewGSServer(address string) (*GSServer, error) {
	s := new(GSServer)
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		fmt.Printf("tls init failed %s\n", err)
		return nil, err
	}
	s.config = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	s.address = address
	s.conns = make(map[string]*tls.Conn)

	return s, nil
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
		connUnenc, err := listener.Accept()
		if err != nil {
			fmt.Printf("server: accept %s", err)
			break
		}
		conn := tls.Server(connUnenc, s.config)
		s.m.Lock()
		s.conns[conn.RemoteAddr().String()] = conn
		s.m.Unlock()
		fmt.Printf("GSServer : accepted from %s", conn.RemoteAddr())
		go s.handleConnection(conn)

	}
	return nil
}

// HandleConnection : handler for the socket
func (s *GSServer) handleConnection(conn *tls.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()
	remAddr := conn.RemoteAddr().String()

	for {
		fmt.Print("Receive command ")
		msgType, err := rw.ReadString('\n')
		switch {
		case err == io.EOF:
			fmt.Println("Reached EOF - close this connection.\n ---")
			s.m.Lock()
			delete(s.conns, remAddr)
			s.m.Unlock()
			return
		case err != nil:
			fmt.Println("Failed to read:", err.Error())
			s.m.Lock()
			delete(s.conns, remAddr)
			s.m.Unlock()
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
				return
			}
			fmt.Printf("Event : \n%#v\n", event)
		case "cmd":
			var command msg.Command
			dec := gob.NewDecoder(rw)
			err := dec.Decode(&command)
			if err != nil {
				fmt.Println("Error decoding Command data ", err)
				s.m.Lock()
				delete(s.conns, remAddr)
				s.m.Unlock()
				return
			}
			fmt.Printf("Event : \n%#v\n", command)
		case "rep":
			var reply msg.Reply
			dec := gob.NewDecoder(rw)
			err := dec.Decode(&reply)
			if err != nil {
				fmt.Println("Error decoding Reply data ", err)
				s.m.Lock()
				delete(s.conns, remAddr)
				s.m.Unlock()
				return
			}
			fmt.Printf("Event : \n%#v\n", reply)
		case "cfg":
			config, err := rw.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading config string", err)
				s.m.Lock()
				delete(s.conns, remAddr)
				s.m.Unlock()
				return
			}
			fmt.Printf("Config: \n\t%s\n", config)
		default:
			fmt.Printf("%s msg type is not implemented\n", msgType)
			s.m.Lock()
			delete(s.conns, remAddr)
			s.m.Unlock()
			return
		}
	}
	/* old loop
	for {
		fmt.Print("Waiting data...")
		msg, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read:", err.Error())
			s.m.Lock()
			delete(s.conns, remAddr)
			s.m.Unlock()
			break
		}

		fmt.Printf("Received and echoing %s", msg)
		_, err = rw.WriteString(msg)
		if err != nil {
			fmt.Println("Failed to write:", err.Error())
			s.m.Lock()
			delete(s.conns, remAddr)
			s.m.Unlock()
			break
		}
		fmt.Printf("successed in writing !")
	}
	*/
}

func server(address string) {
	test, err := NewGSServer(address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	} else {
		test.Run()
	}
}
