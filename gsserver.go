package main

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"errors"
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
	rw         *bufio.ReadWriter
	m          sync.RWMutex
}

//add : Add a new connection from a client
func (s *GSServer) add(conn *tls.Conn) (*GSServerConn, error) {
	c := new(GSServerConn)
	c.socket = conn
	c.gsServer = s
	c.address = conn.RemoteAddr().String()
	c.localName = s.name
	s.m.Lock()
	s.conns[c.address] = c
	s.m.Unlock()
	c.rw = new(bufio.ReadWriter)
	//go conn.Run()
	return c, nil
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

func (s *GSServerConn) getMsg(data interface{}) error {
	s.m.Lock()
	fmt.Printf("NewDecoder")
	dec := gob.NewDecoder(s.rw)
	fmt.Printf("Decode")
	err := dec.Decode(&data)
	s.m.Unlock()
	if err != nil {
		fmt.Println("Error decoding received message ", err)
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
	}
	return err
}

func (s *GSServerConn) receiveMsg(msgType string) (interface{}, error) {
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
		fmt.Printf("%s msg type is not implemented\n", msgType)
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
		return nil, errors.New("non implemented message " + msgType)
	}
	s.m.Lock()
	dec := gob.NewDecoder(s.rw)
	err := dec.Decode(&data)
	if err != nil {
		fmt.Println("Error decoding received message ", err)
		s.gsServer.m.Lock()
		delete(s.gsServer.conns, s.address)
		s.gsServer.m.Unlock()
	}
	s.m.Unlock()
	fmt.Printf("Receive %s: \n%#v\n", msgType, data)
	return &data, err
}

// Run : handler for the connection
func (s *GSServerConn) Run() {
	s.rw = bufio.NewReadWriter(bufio.NewReader(s.socket), bufio.NewWriter(s.socket))

	for {
		fmt.Print("Receive command ")
		msgType, err := s.rw.ReadString('\n')
		switch {
		case err == io.EOF:
			fmt.Println("Reached EOF - close this connection.\n ---")
			s.gsServer.m.Lock()
			delete(s.gsServer.conns, s.address)
			s.gsServer.m.Unlock()
			return
		case err != nil:
			fmt.Println("Failed to read:", err.Error())
			s.gsServer.m.Lock()
			delete(s.gsServer.conns, s.address)
			s.gsServer.m.Unlock()
			return
		}
		msgType = strings.Trim(msgType, "\n")
		fmt.Println("'" + msgType + "'")
		fmt.Println("Receiving data part of message")
		_, err = s.receiveMsg(msgType)
		if err != nil {
			return
		}
	}
}

func server(name string, address string) {
	s, err := NewGSServer(name, address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	}
	<-s.done

}
