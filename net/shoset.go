package net

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

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
	//	id          string
	connsByAddr  map[string]*ShosetConn            // ensemble des connexions
	connsByName  map[string]map[string]*ShosetConn // connexions par nom logique
	conssByType  map[string]map[string]*ShosetConn // connexions par type
	connsJoin    map[string]*ShosetConn            // connexions nécessaires au join (non utilisées en dehors du join)
	brothers     map[string]bool                   // "freres" au sens large (ex: toutes les instances de connecteur reliées à un même aggregateur)
	nameBrothers map[string]bool                   // "freres" ayant un même nom logique (ex: instances d'un même connecteur)
	lName        string                            // Nom logique de la shoset
	ShosetType   string                            // Type logique de la shoset
	bindAddr     string                            // Adresse sur laquelle la shoset est bindée

	// map des queues par type de message (enregistrées via RegisterMessageBehaviors)
	queue map[string]*msg.Queue

	// map des fonctions de gestion des messages par type (enregistrées via RegisterMessageBehaviors)
	handle map[string]func(*ShosetConn) error

	// map des fonctions d'envoi de message (enregistrées via RegisterMessageBehaviors)
	send map[string]func(*Shoset, msg.Message)

	// map des fonctions d'attente de message (enregistrées via RegisterMessageBehaviors)
	// les fonctions d'attente sont synchrones, à charge à l'utilisateur de gérer l'asynchronisme selon le language qu'il utilise
	// une fonction d'attente possède 3 arguments :
	//   - un iterateur a appeler pour chaque nouevel élément
	//   - un filtre sur le message (défini par une map de strings)
	// 	 - un timeout après lequel la fonction retourne nil si aucun message n'est arrivé.
	wait map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message

	// configuration TLS
	tlsConfig   *tls.Config
	tlsServerOK bool

	// synchronisation des goroutines
	Done chan bool
	m    sync.RWMutex
	ma   sync.RWMutex
	mj   sync.RWMutex
}

var certPath = "./certs/cert.pem"
var keyPath = "./certs/key.pem"

// NewShoset : constructor
func NewShoset(lName, ShosetType string) *Shoset {
	// Creation
	c := new(Shoset)

	// Initialisation
	c.lName = lName
	c.ShosetType = ShosetType
	c.connsByAddr = make(map[string]*ShosetConn)
	c.conssByType = make(map[string]map[string]*ShosetConn)
	c.connsByName = make(map[string]map[string]*ShosetConn)
	c.connsJoin = make(map[string]*ShosetConn)
	c.brothers = make(map[string]bool)
	c.nameBrothers = make(map[string]bool)

	c.queue = make(map[string]*msg.Queue)
	c.handle = make(map[string]func(*ShosetConn) error)
	c.send = make(map[string]func(*Shoset, msg.Message))
	c.wait = make(map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message)

	// Enregistrement des handlers par type de message  (fonctions de gestion, d'envoi et de reception des messages)
	c.RegisterMessageBehaviors("cfg", HandleConfig, SendConfig, WaitConfig)    // messages de type configuration
	c.RegisterMessageBehaviors("evt", HandleEvent, SendEvent, WaitEvent)       // messages de type événement
	c.RegisterMessageBehaviors("cmd", HandleCommand, SendCommand, WaitCommand) // messages de type commande

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
func (c *Shoset) RegisterMessageBehaviors(
	msgType string,
	handle func(*ShosetConn) error,
	send func(*Shoset, msg.Message),
	wait func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message) {
	c.queue[msgType] = msg.NewQueue()
	c.handle[msgType] = handle
	c.send[msgType] = send
	c.wait[msgType] = wait
}

// SetBrother :
func (c *Shoset) SetBrother(brother string) {
	c.m.Lock()
	c.brothers[brother] = true
	c.m.Unlock()
}

// SetNameBrother :
func (c *Shoset) SetNameBrother(nameBrother string) {
	c.m.Lock()
	c.nameBrothers[nameBrother] = true
	c.m.Unlock()
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

// GetConnByAddr :
func (c *Shoset) GetConnByAddr(addr string) *ShosetConn {
	c.ma.Lock()
	defer c.ma.Unlock()
	return c.connsByAddr[addr]
}

// String :
func (c *Shoset) String() string {
	str := fmt.Sprintf("Shoset { lName: %s, bindAddr: %s, type: %s, brothers %#v, nameBrothers %#v, joinConns %#v\n", c.lName, c.bindAddr, c.ShosetType, c.brothers, c.nameBrothers, c.connsJoin)
	for k, conn := range c.connsByAddr {
		str += fmt.Sprintf(" - [%s] %s\n", k, conn.String())
	}
	str += fmt.Sprintf("}\n")
	return str
}

// Queue :
func (c *Shoset) Queue(msgType string) *msg.Queue {
	return c.queue[msgType]
}

// Handle :
func (c *Shoset) Handle(msgType string) func(*ShosetConn) error {
	return c.handle[msgType]
}

// Send :
func (c *Shoset) Send(msgType string) func(*Shoset, msg.Message) {
	return c.send[msgType]
}

// Wait :
func (c *Shoset) Wait(msgType string) func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message {
	return c.wait[msgType]
}

//NewHandshake : Build a config Message
func (c *Shoset) NewHandshake() *msg.Config {
	return msg.NewHandshake(c.bindAddr, c.lName, c.ShosetType)
}

//NewCfgOut : Build a config Message
func (c *Shoset) NewCfgOut() *msg.Config {
	var bros []string
	c.ma.Lock()
	for _, conn := range c.connsByAddr {
		if conn.dir == "out" {
			bros = append(bros, conn.addr)
		}
	}
	c.ma.Unlock()
	return msg.NewConns("out", bros)
}

//NewCfgIn : Build a config Message
func (c *Shoset) NewCfgIn() *msg.Config {
	var bros []string
	c.m.Lock()
	for addr := range c.brothers {
		bros = append(bros, addr)
	}
	c.m.Unlock()
	return msg.NewConns("in", bros)
}

//NewInstanceMessage : Build a config Message
func (c *Shoset) NewInstanceMessage(address, lName, ShosetType string) *msg.Config {
	return msg.NewInstance(address, lName, ShosetType)
}

//NewConnectToMessage : Build a config Message
func (c *Shoset) NewConnectToMessage(address string) *msg.Config {
	return msg.NewConnectTo(address)
}

//Connect : Connect to another Shoset
func (c *Shoset) Connect(address string) (*ShosetConn, error) {
	conn := new(ShosetConn)
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

//Join : Join to group of Shosets and duplicate in and out connexions
func (c *Shoset) Join(address string) (*ShosetConn, error) {

	c.mj.Lock()
	conn, ok := c.connsJoin[address]
	c.mj.Unlock()
	if ok {
		return conn, nil
	}
	if address == c.bindAddr {
		return nil, nil
	}

	conn = new(ShosetConn)
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

func (c *Shoset) deleteConn(connAddr string) {
	c.m.Lock()
	c.ma.Lock()
	conn := c.connsByAddr[connAddr]
	if conn != nil {
		lName := conn.name
		if c.connsByName[lName] != nil {
			delete(c.connsByName[lName], connAddr)
		}

		ShosetType := conn.ShosetType
		if c.conssByType[ShosetType] != nil {
			delete(c.conssByType[ShosetType], connAddr)
		}

		delete(c.connsByAddr, connAddr)
	}
	c.m.Unlock()
	c.ma.Unlock()
}

// SetConnJoin :
func (c *Shoset) SetConnJoin(connAddr string, conn *ShosetConn) {
	if conn != nil {
		c.mj.Lock()
		c.connsJoin[connAddr] = conn
		c.mj.Unlock()
	}
}

// SetConn :
func (c *Shoset) SetConn(connAddr, connType string, conn *ShosetConn) {
	if conn != nil {
		c.m.Lock()
		c.ma.Lock()
		c.connsByAddr[connAddr] = conn

		if c.conssByType[connType] == nil {
			c.conssByType[connType] = make(map[string]*ShosetConn)
		}
		c.conssByType[connType][connAddr] = conn
		c.m.Unlock()
		c.ma.Unlock()
		c.SetConnByName(conn)
	}
}

// SetConnByName :
func (c *Shoset) SetConnByName(conn *ShosetConn) {
	c.m.Lock()
	defer c.m.Unlock()
	lName := conn.name
	if lName != "" {
		if c.connsByName[lName] == nil {
			c.connsByName[lName] = make(map[string]*ShosetConn)
		}
		c.connsByName[lName][conn.addr] = conn
	}
}

// SendCfgToBrothers :
func (c *Shoset) SendCfgToBrothers(currentConn *ShosetConn) {
	currentAddr := currentConn.GetBindAddr()
	currentName := currentConn.GetName()
	cfgNameBrothers := msg.NewNameBrothers([]string{currentAddr})
	oldNameBrothers := []string{}
	c.m.Lock()
	for _, conn := range c.connsByName[currentName] {
		if conn.GetDir() == "in" && conn.bindAddr != currentAddr {
			conn.SendMessage(cfgNameBrothers)
			oldNameBrothers = append(oldNameBrothers, conn.bindAddr)
		}
	}
	c.m.Unlock()
	if len(oldNameBrothers) > 0 {
		currentConn.SendMessage(msg.NewNameBrothers(oldNameBrothers))
	}
}

// SendConnectToNameBrothers :
func (c *Shoset) SendConnectToNameBrothers(conn *ShosetConn) {
	if len(c.nameBrothers) > 0 {
		c.m.Lock()
		defer c.m.Unlock()
		for bro := range c.nameBrothers {
			conn.SendMessage(msg.NewConnectTo(bro))
		}
	}
}

// SendConnectToInConnsByAddr :
func (c *Shoset) SendConnectToInConnsByAddr(addr string) {
	c.ma.Lock()
	defer c.ma.Unlock()
	if !c.nameBrothers[addr] {
		for _, conn := range c.connsByAddr {
			if conn.GetDir() == "in" {
				conn.SendMessage(msg.NewConnectTo(addr))
			}
		}
		c.nameBrothers[addr] = true
	}
}

// SendMsgToConnsByAddr :
func (c *Shoset) SendMsgToConnsByAddr(message msg.Message) {
	c.ma.Lock()
	defer c.ma.Unlock()
	for _, conn := range c.connsByAddr {
		conn.SendMessage(message)
	}
}

// SendCfgInToInConnsByAddr :
func (c *Shoset) SendCfgInToInConnsByAddr(currentConn *ShosetConn) {
	c.ma.Lock()
	defer c.ma.Unlock()
	cfgIn := c.NewCfgIn()
	for _, conn := range c.connsByAddr {
		if conn.GetDir() == "in" {
			currentConn.SendMessage(cfgIn)
		}
	}
}

// SendNewMemberToConnsJoin :
func (c *Shoset) SendNewMemberToConnsJoin(addr string) {
	thisOne := c.bindAddr
	cfgNewMember := msg.NewCfgMember(addr)
	c.mj.Lock()
	for connAddr, conn := range c.connsJoin {
		if connAddr != addr && connAddr != thisOne {
			conn.SendMessage(cfgNewMember)
		}
	}
	c.mj.Unlock()
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
