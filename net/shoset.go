package net

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"

	"shoset/msg"
)

// MessageHandlers interface
type MessageHandlers interface {
	Handle(*ShosetConn) error
	SendConn(*ShosetConn, *msg.Message)
	Send(*Shoset, *msg.Message)
	Wait(*Shoset, *msg.Iterator, string, int) *msg.Message
}

//Shoset :
type Shoset struct {
	Context 	map[string]interface{} //TOTO
	
	//	id          string
	ConnsByAddr  *MapSafeConn    // map[string]*ShosetConn    ensemble des connexions
	ConnsByName  *MapSafeMapConn // map[string]map[string]*ShosetConn   connexions par nom logique
	ConnsByType  *MapSafeMapConn // map[string]map[string]*ShosetConn   connexions par type
	ConnsJoin    *MapSafeConn    // map[string]*ShosetConn    connexions nécessaires au join (non utilisées en dehors du join)
	Brothers     *MapSafeBool    // map[string]bool  "freres" au sens large (ex: toutes les instances de connecteur reliées à un même aggregateur)
	NameBrothers *MapSafeBool    // map[string]bool  "freres" ayant un même nom logique (ex: instances d'un même connecteur)

	lName      string // Nom logique de la shoset
	ShosetType string // Type logique de la shoset
	bindAddr   string // Adresse sur laquelle la shoset est bindée

	// Dictionnaire des queues de message (par type de message)
	Queue map[string]*msg.Queue

	Get    map[string]func(*ShosetConn) (msg.Message, error)
	Handle map[string]func(*ShosetConn, msg.Message) error
	Send   map[string]func(*Shoset, msg.Message)
	Wait   map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message

	// configuration TLS
	tlsConfig   *tls.Config
	tlsServerOK bool

	// synchronisation des goroutines
	Done chan bool
}

var certPath = "./certs/cert.pem"
var keyPath = "./certs/key.pem"

// NewShoset : constructor
func NewShoset(lName, ShosetType string) *Shoset {
	// Creation
	c := new(Shoset)

	c.Context = make(map[string]interface{})

	// Initialisation
	c.lName = lName
	c.ShosetType = ShosetType
	c.ConnsByAddr = NewMapSafeConn()
	c.ConnsByName = NewMapSafeMapConn()
	c.ConnsByType = NewMapSafeMapConn()
	c.ConnsJoin = NewMapSafeConn()
	c.Brothers = NewMapSafeBool()
	c.NameBrothers = NewMapSafeBool()

	c.Queue = make(map[string]*msg.Queue)
	c.Get = make(map[string]func(*ShosetConn) (msg.Message, error))
	c.Handle = make(map[string]func(*ShosetConn, msg.Message) error)
	c.Send = make(map[string]func(*Shoset, msg.Message))
	c.Wait = make(map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message)

	c.Queue["cfglink"] = msg.NewQueue()
	c.Get["cfglink"] = GetConfigLink
	c.Handle["cfglink"] = HandleConfigLink

	c.Queue["cfgjoin"] = msg.NewQueue()
	c.Get["cfgjoin"] = GetConfigJoin
	c.Handle["cfgjoin"] = HandleConfigJoin

	c.Queue["evt"] = msg.NewQueue()
	c.Get["evt"] = GetEvent
	c.Handle["evt"] = HandleEvent
	c.Send["evt"] = SendEvent
	c.Wait["evt"] = WaitEvent

	c.Queue["cmd"] = msg.NewQueue()
	c.Get["cmd"] = GetCommand
	c.Handle["cmd"] = HandleCommand
	c.Send["cmd"] = SendCommand
	c.Wait["cmd"] = WaitCommand

	// Configuration TLS
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

// GetBindAddr :
func (c *Shoset) GetBindAddr() string {
	return c.bindAddr
}

// GetName :
func (c *Shoset) GetName() string {
	return c.lName
}

// GetShosetType :
func (c *Shoset) GetShosetType() string { return c.ShosetType }

// String :
func (c *Shoset) String() string {
	descr := fmt.Sprintf("Shoset { lName: %s, bindAddr: %s, type: %s, brothers %#v, nameBrothers %#v, joinConns %#v\n", c.lName, c.bindAddr, c.ShosetType, c.Brothers, c.NameBrothers, c.ConnsJoin)
	c.ConnsByAddr.Iterate(
		func(key string, val *ShosetConn) {
			descr = fmt.Sprintf("%s - [%s] %s\n", descr, key, val.String())
		})
	descr += "%s}\n"
	return descr
}

//Link : Link to another Shoset
func (c *Shoset) Link(address string) (*ShosetConn, error) {
	conn := new(ShosetConn)
	conn.ch = c
	conn.dir = "out"
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}
	conn.addr = ipAddress
	conn.brothers = make(map[string]bool)
	go conn.runOutConn(conn.addr)
	return conn, nil
}

//Join : Join to group of Shosets and duplicate in and out connexions
func (c *Shoset) Join(address string) (*ShosetConn, error) {

	exists := c.ConnsJoin.Get(address)
	if exists != nil {
		return exists, nil
	}
	if address == c.bindAddr {
		return nil, nil
	}

	conn := new(ShosetConn)
	conn.ch = c
	conn.dir = "out"
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}
	conn.addr = ipAddress
	conn.bindAddr = ipAddress
	conn.brothers = make(map[string]bool)
	go conn.runJoinConn()
	return conn, nil
}

func (c *Shoset) deleteConn(connAddr string) {
	conn := c.ConnsByAddr.Get(connAddr)
	if conn != nil {
		c.ConnsByName.Delete(conn.name, connAddr)
		c.ConnsByType.Delete(conn.ShosetType, connAddr)
		c.ConnsByAddr.Delete(connAddr)
	}
}

// SetConn :
func (c *Shoset) SetConn(connAddr, connType string, conn *ShosetConn) {
	if conn != nil {
		c.ConnsByAddr.Set(connAddr, conn)
		c.ConnsByType.Set(connType, conn.addr, conn)
		c.ConnsByName.Set(conn.name, conn.addr, conn)
	}
}

//Bind : Connect to another Shoset
func (c *Shoset) Bind(address string) error {
	if c.bindAddr != "" {
		fmt.Println("Shoset already bound")
		return errors.New("Shoset already bound")
	}
	if c.tlsServerOK == false {
		fmt.Println("TLS configuration not OK (certificate not found / loaded)")
		return errors.New("TLS configuration not OK (certificate not found / loaded)")
	}
	ipAddress, err := GetIP(address)
	if err != nil {
		return err
	}
	c.bindAddr = ipAddress
	//fmt.Printf("Bind : handleBind adress %s", ipAddress)
	go c.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (c *Shoset) handleBind() error {
	listener, err := net.Listen("tcp", c.bindAddr)
	if err != nil {
		fmt.Println("Failed to bind:", err.Error())
		return err
	}
	defer listener.Close()

	for {
		unencConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("serverShoset accept error: %s", err)
			break
		}
		tlsConn := tls.Server(unencConn, c.tlsConfig)
		conn, _ := c.inboudConn(tlsConn)
		//fmt.Printf("Shoset : accepted from %s", conn.addr)
		go conn.runInConn()
	}
	return nil
}

//inboudConn : Add a new connection from a client
func (c *Shoset) inboudConn(tlsConn *tls.Conn) (*ShosetConn, error) {
	conn := new(ShosetConn)
	conn.socket = tlsConn
	conn.dir = "in"
	conn.ch = c
	conn.addr = tlsConn.RemoteAddr().String()
	//c.SetConn(conn.addr, conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	return conn, nil
}
