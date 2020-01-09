package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	//	uuid "github.com/kjk/betterguid"

	"./msg"
)

//Chaussette : client gandalf Socket
type Chaussette struct {
	//	id          string
	connsByAddr map[string]*ChaussetteConn
	connsByName map[string]map[string]*ChaussetteConn
	brothers    map[string]bool
	logicalName string
	done        chan bool
	bindAddress string
	m           sync.RWMutex
	qEvents     *msg.Queue
	qCommands   *msg.Queue
	qReplies    *msg.Queue
	qConfigs    *msg.Queue
	tlsConfig   *tls.Config
	tlsServerOK bool
}

func (c *Chaussette) String() string {
	str := fmt.Sprintf("Chaussette{ logicalName: %s, bindAddress: %s, brothers %#v\n", c.logicalName, c.bindAddress, c.brothers)
	for _, conn := range c.connsByAddr {
		str += fmt.Sprintf(" - %s\n", conn.String())
	}
	str += fmt.Sprintf("\n")
	return str
}

var certPath = "./cert.pem"
var keyPath = "./key.pem"

// NewChaussette : constructor
func NewChaussette(logicalName string) *Chaussette {
	c := new(Chaussette)
	//	c.id = uuid.New()
	c.logicalName = logicalName
	c.connsByAddr = make(map[string]*ChaussetteConn)
	c.connsByName = make(map[string]map[string]*ChaussetteConn)
	c.brothers = make(map[string]bool)
	c.qEvents = msg.NewQueue()
	c.qCommands = msg.NewQueue()
	c.qReplies = msg.NewQueue()
	c.qConfigs = msg.NewQueue()

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil { // only client in insecure mode
		fmt.Println("Unable to Load certificate")
		c.tlsConfig = &tls.Config{InsecureSkipVerify: true}
		c.tlsServerOK = false
	} else {
		c.tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		c.tlsServerOK = true
	}
	return c
}

//NewInstanceMessage : Build a config Message
func (c *Chaussette) NewInstanceMessage(address string, logicalName string) *msg.Config {
	return msg.NewInstance(address, logicalName)
}

//NewConnectToMessage : Build a config Message
func (c *Chaussette) NewConnectToMessage(address string) *msg.Config {
	return msg.NewConnectTo(address)
}

//Connect : Connect to another Chaussette
func (c *Chaussette) Connect(address string) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.chaussette = c
	conn.direction = "out"
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := getIP(address)
	if err != nil {
		return nil, err
	}
	conn.address = ipAddress
	go conn.runOutConn(conn.address)
	return conn, nil
}

func (c *Chaussette) deleteConn(connAddr string) {
	c.m.Lock()
	conn := c.connsByAddr[connAddr]
	if conn != nil {
		logicalName := conn.remoteLogicalName
		if c.connsByName[logicalName] != nil {
			delete(c.connsByName[logicalName], connAddr)
		}
	}
	delete(c.connsByAddr, connAddr)
	c.m.Unlock()
}

func (c *Chaussette) setConn(connAddr string, conn *ChaussetteConn) {
	if conn != nil {
		c.m.Lock()
		c.connsByAddr[connAddr] = conn
		logicalName := conn.remoteLogicalName
		if logicalName != "" {
			if c.connsByName[logicalName] == nil {
				c.connsByName[logicalName] = make(map[string]*ChaussetteConn)
			}
			c.connsByName[logicalName][connAddr] = conn
		}
		c.m.Unlock()
	}
}

//Bind : Connect to another Chaussette
func (c *Chaussette) Bind(address string) error {
	if c.bindAddress != "" {
		fmt.Println("Chaussette already bound")
		return errors.New("Chaussette already bound")
	}
	if c.tlsServerOK == false {
		fmt.Println("TLS configuration not OK (certificate not found / loaded)")
		return errors.New("TLS configuration not OK (certificate not found / loaded)")
	}
	ipAddress, err := getIP(address)
	if err != nil {
		return err
	}
	c.bindAddress = ipAddress
	fmt.Printf("Bind : handleBind adress %s", ipAddress)
	go c.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (c *Chaussette) handleBind() error {
	listener, err := net.Listen("tcp", c.bindAddress)
	if err != nil {
		fmt.Println("Failed to bind:", err.Error())
		return err
	}
	defer listener.Close()

	for {
		unencConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("serverChaussette accept error: %s", err)
			break
		}
		tlsConn := tls.Server(unencConn, c.tlsConfig)
		conn, _ := c.inboudConn(tlsConn)
		fmt.Printf("Chaussette : accepted from %s", conn.address)
		go conn.runInConn()
	}
	return nil
}

//inboudConn : Add a new connection from a client
func (c *Chaussette) inboudConn(tlsConn *tls.Conn) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.socket = tlsConn
	conn.direction = "in"
	conn.chaussette = c
	conn.address = tlsConn.RemoteAddr().String()
	c.setConn(conn.address, conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	return conn, nil
}
