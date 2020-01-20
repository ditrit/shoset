package net

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	//	uuid "github.com/kjk/betterguid"

	"../msg"
)

// MessageHandlers interface
type MessageHandlers interface {
	Handle(*ChaussetteConn) error
	SendConn(*ChaussetteConn, *msg.Message)
	Send(*Chaussette, *msg.Message)
	Wait(*Chaussette, *msg.Iterator, string, int) *msg.Message
}

//Chaussette :
type Chaussette struct {
	//	id          string
	connsByAddr  map[string]*ChaussetteConn            // ensemble des connexions
	connsByName  map[string]map[string]*ChaussetteConn // connexions par nom logique
	connsJoin    map[string]*ChaussetteConn            // connexions nécessaires au join (non utilisées en dehors du join)
	brothers     map[string]bool                       // "freres" au sens large (ex: toutes les instances de connecteur reliées à un même aggregateur)
	nameBrothers map[string]bool                       // "freres" ayant un même nom logique (ex: instances d'un même connecteur)
	lName        string                                // Nom logique de la chaussette
	bindAddr     string                                // Adresse sur laquelle la chaussette est bindée

	// map des queues par type de message (enregistrées via RegisterMessageBehaviors)
	queue map[string]*msg.Queue

	// map des fonctions de gestion des messages par type (enregistrées via RegisterMessageBehaviors)
	handle map[string]func(*ChaussetteConn) error

	// map des fonctions d'envoi de message (enregistrées via RegisterMessageBehaviors)
	send map[string]func(*Chaussette, msg.Message)

	// map des fonctions d'attente de message (enregistrées via RegisterMessageBehaviors)
	// les fonctions d'attente sont synchrones, à charge à l'utilisateur de gérer l'asynchronisme selon le language qu'il utilise
	// une fonction d'attente possède 3 arguments :
	//   - un iterateur a appeler pour chaque nouevel élément
	//   - un filtre sur le message (défini par une map de strings)
	// 	 - un timeout après lequel la fonction retourne nil si aucun message n'est arrivé.
	wait map[string]func(*Chaussette, *msg.Iterator, map[string]string, int) *msg.Message

	// configuration TLS
	tlsConfig   *tls.Config
	tlsServerOK bool

	// synchronisation des goroutines
	Done chan bool
	m    sync.RWMutex
}

var certPath = "./certs/cert.pem"
var keyPath = "./certs/key.pem"

// NewChaussette : constructor
func NewChaussette(lName string) *Chaussette {
	// Creation
	c := new(Chaussette)

	// Initialisation
	c.lName = lName
	c.connsByAddr = make(map[string]*ChaussetteConn)
	c.connsByName = make(map[string]map[string]*ChaussetteConn)
	c.connsJoin = make(map[string]*ChaussetteConn)
	c.brothers = make(map[string]bool)
	c.nameBrothers = make(map[string]bool)

	c.queue = make(map[string]*msg.Queue)
	c.handle = make(map[string]func(*ChaussetteConn) error)
	c.send = make(map[string]func(*Chaussette, msg.Message))
	c.wait = make(map[string]func(*Chaussette, *msg.Iterator, map[string]string, int) *msg.Message)

	// Enregistrement des handlers par type de message  (fonctions de gestion, d'envoi et de reception des messages)
	c.RegisterMessageBehaviors("cfg", HandleConfig, SendConfig, WaitConfig)    // messages de type configuration
	c.RegisterMessageBehaviors("evt", HandleEvent, SendEvent, WaitEvent)       // messages de type événement
	c.RegisterMessageBehaviors("cmd", HandleCommand, SendCommand, WaitCommand) // messages de type commande
	c.RegisterMessageBehaviors("rep", HandleReply, SendReply, WaitReply)       // messages de type réponse à une commande

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

// RegisterMessageBehaviors :
func (c *Chaussette) RegisterMessageBehaviors(
	msgType string,
	handle func(*ChaussetteConn) error,
	send func(*Chaussette, msg.Message),
	wait func(*Chaussette, *msg.Iterator, map[string]string, int) *msg.Message) {
	c.queue[msgType] = msg.NewQueue()
	c.handle[msgType] = handle
	c.send[msgType] = send
	c.wait[msgType] = wait
}

// GetBrothers :
func (c *Chaussette) GetBrothers() map[string]bool {
	return c.brothers
}

// SetBrother :
func (c *Chaussette) SetBrother(brother string) {
	c.brothers[brother] = true
}

// GetNameBrothers :
func (c *Chaussette) GetNameBrothers() map[string]bool {
	return c.nameBrothers
}

// InNameBrothers :
func (c *Chaussette) InNameBrothers(addr string) bool {
	return c.nameBrothers[addr]
}

// InConnsJoin :
func (c *Chaussette) InConnsJoin(addr string) bool {
	_, ok := c.connsJoin[addr]
	return ok
}

// SetNameBrother :
func (c *Chaussette) SetNameBrother(nameBrother string) {
	c.nameBrothers[nameBrother] = true
}

// GetBindAddr :
func (c *Chaussette) GetBindAddr() string {
	return c.bindAddr
}

// GetName :
func (c *Chaussette) GetName() string {
	return c.lName
}

// GetConnsByAddr :
func (c *Chaussette) GetConnsByAddr() map[string]*ChaussetteConn {
	return c.connsByAddr
}

// GetConnsByName :
func (c *Chaussette) GetConnsByName() map[string]map[string]*ChaussetteConn {
	return c.connsByName
}

// GetConnsJoin :
func (c *Chaussette) GetConnsJoin() map[string]*ChaussetteConn {
	return c.connsJoin
}

// String :
func (c *Chaussette) String() string {
	str := fmt.Sprintf("Chaussette{ lName: %s, bindAddr: %s, brothers %#v, nameBrothers %#v, joinConns %#v\n", c.lName, c.bindAddr, c.brothers, c.nameBrothers, c.connsJoin)
	for k, conn := range c.connsByAddr {
		str += fmt.Sprintf(" - [%s] %s\n", k, conn.String())
	}
	str += fmt.Sprintf("\n")
	return str
}

// FQueue :
func (c *Chaussette) FQueue(msgType string) *msg.Queue {
	return c.queue[msgType]
}

// FHandle :
func (c *Chaussette) FHandle(msgType string) func(*ChaussetteConn) error {
	return c.handle[msgType]
}

// FSend :
func (c *Chaussette) FSend(msgType string) func(*Chaussette, msg.Message) {
	return c.send[msgType]
}

// FWait :
func (c *Chaussette) FWait(msgType string) func(*Chaussette, *msg.Iterator, map[string]string, int) *msg.Message {
	return c.wait[msgType]
}

//NewHandshake : Build a config Message
func (c *Chaussette) NewHandshake() *msg.Config {
	return msg.NewHandshake(c.bindAddr, c.lName)
}

//NewCfgOut : Build a config Message
func (c *Chaussette) NewCfgOut() *msg.Config {
	var bros []string
	for _, conn := range c.GetConnsByAddr() {
		if conn.dir == "out" {
			bros = append(bros, conn.addr)
		}
	}
	return msg.NewConns("out", bros)
}

//NewCfgIn : Build a config Message
func (c *Chaussette) NewCfgIn() *msg.Config {
	var bros []string
	for addr := range c.brothers {
		bros = append(bros, addr)
	}
	return msg.NewConns("in", bros)
}

//NewInstanceMessage : Build a config Message
func (c *Chaussette) NewInstanceMessage(address string, lName string) *msg.Config {
	return msg.NewInstance(address, lName)
}

//NewConnectToMessage : Build a config Message
func (c *Chaussette) NewConnectToMessage(address string) *msg.Config {
	return msg.NewConnectTo(address)
}

//Connect : Connect to another Chaussette
func (c *Chaussette) Connect(address string) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.ch = c
	conn.dir = "out"
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := getIP(address)
	if err != nil {
		return nil, err
	}
	conn.addr = ipAddress
	conn.brothers = make(map[string]bool)
	go conn.runOutConn(conn.addr)
	return conn, nil
}

//Join : Join to group of Chaussettes and duplicate in and out connexions
func (c *Chaussette) Join(address string) (*ChaussetteConn, error) {

	conn := new(ChaussetteConn)
	conn.ch = c
	conn.dir = "out"
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := getIP(address)
	if err != nil {
		return nil, err
	}
	conn.addr = ipAddress
	conn.bindAddr = ipAddress
	conn.brothers = make(map[string]bool)
	go conn.runJoinConn()
	return conn, nil
}

func (c *Chaussette) deleteConn(connAddr string) {
	c.m.Lock()
	conn := c.connsByAddr[connAddr]
	if conn != nil {
		lName := conn.name
		if c.connsByName[lName] != nil {
			delete(c.connsByName[lName], connAddr)
		}
	}
	delete(c.connsByAddr, connAddr)
	c.m.Unlock()
}

// SetConnJoin :
func (c *Chaussette) SetConnJoin(connAddr string, conn *ChaussetteConn) {
	if conn != nil {
		c.m.Lock()
		c.connsJoin[connAddr] = conn
		c.m.Unlock()
	}
}

// SetConn :
func (c *Chaussette) SetConn(connAddr string, conn *ChaussetteConn) {
	if conn != nil {
		c.m.Lock()
		c.connsByAddr[connAddr] = conn
		lName := conn.name
		if lName != "" {
			if c.connsByName[lName] == nil {
				c.connsByName[lName] = make(map[string]*ChaussetteConn)
			}
			c.connsByName[lName][connAddr] = conn
		}
		c.m.Unlock()
	}
}

//Bind : Connect to another Chaussette
func (c *Chaussette) Bind(address string) error {
	if c.bindAddr != "" {
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
	c.bindAddr = ipAddress
	//fmt.Printf("Bind : handleBind adress %s", ipAddress)
	go c.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (c *Chaussette) handleBind() error {
	listener, err := net.Listen("tcp", c.bindAddr)
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
		//fmt.Printf("Chaussette : accepted from %s", conn.addr)
		go conn.runInConn()
	}
	return nil
}

//inboudConn : Add a new connection from a client
func (c *Chaussette) inboudConn(tlsConn *tls.Conn) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.socket = tlsConn
	conn.dir = "in"
	conn.ch = c
	conn.addr = tlsConn.RemoteAddr().String()
	//c.SetConn(conn.addr, conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	return conn, nil
}
