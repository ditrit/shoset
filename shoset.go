package shoset

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ditrit/shoset/msg"
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
	Context map[string]interface{} //TOTO

	//	id          string
	ConnsByAddr  *MapSafeConn    // map[string]*ShosetConn    ensemble des connexions
	ConnsByName  *MapSafeMapConn // map[string]map[string]*ShosetConn   connexions par nom logique
	ConnsByType  *MapSafeMapConn // map[string]map[string]*ShosetConn   connexions par type
	ConnsJoin    *MapSafeConn    // map[string]*ShosetConn    connexions nécessaires au join (non utilisées en dehors du join)
	ConnsBye     *MapSafeConn    // map[string]*ShosetConn    utilisée uniquement comme liste temporaire des connections pour Bye
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
	sh := new(Shoset)

	sh.Context = make(map[string]interface{})

	// Initialisation
	sh.lName = lName
	sh.ShosetType = ShosetType
	sh.ConnsByAddr = NewMapSafeConn()
	sh.ConnsByName = NewMapSafeMapConn()
	sh.ConnsByType = NewMapSafeMapConn()
	sh.ConnsJoin = NewMapSafeConn()
	sh.ConnsBye = NewMapSafeConn()
	sh.Brothers = NewMapSafeBool()
	sh.NameBrothers = NewMapSafeBool()

	sh.Queue = make(map[string]*msg.Queue)
	sh.Get = make(map[string]func(*ShosetConn) (msg.Message, error))
	sh.Handle = make(map[string]func(*ShosetConn, msg.Message) error)
	sh.Send = make(map[string]func(*Shoset, msg.Message))
	sh.Wait = make(map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message)

	sh.Queue["cfglink"] = msg.NewQueue()
	sh.Get["cfglink"] = GetConfigLink
	sh.Handle["cfglink"] = HandleConfigLink

	sh.Queue["cfgjoin"] = msg.NewQueue()
	sh.Get["cfgjoin"] = GetConfigJoin
	sh.Handle["cfgjoin"] = HandleConfigJoin

	sh.Queue["cfgbye"] = msg.NewQueue()
	sh.Get["cfgbye"] = GetConfigBye
	sh.Handle["cfgbye"] = HandleConfigBye

	sh.Queue["evt"] = msg.NewQueue()
	sh.Get["evt"] = GetEvent
	sh.Handle["evt"] = HandleEvent
	sh.Send["evt"] = SendEvent
	sh.Wait["evt"] = WaitEvent

	sh.Queue["cmd"] = msg.NewQueue()
	sh.Get["cmd"] = GetCommand
	sh.Handle["cmd"] = HandleCommand
	sh.Send["cmd"] = SendCommand
	sh.Wait["cmd"] = WaitCommand
	//TODO MOVE TO GANDALF
	sh.Queue["config"] = msg.NewQueue()
	sh.Get["config"] = GetConfig
	sh.Handle["config"] = HandleConfig
	sh.Send["config"] = SendConfig
	sh.Wait["config"] = WaitConfig

	// Configuration TLS
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil { // only client in insecure mode
		fmt.Println("Unable to Load certificate")
		sh.tlsConfig = &tls.Config{InsecureSkipVerify: true}
		sh.tlsServerOK = false
	} else {
		sh.tlsConfig = &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		}
		sh.tlsServerOK = true
	}

	return sh
}

// GetBindAddr :
func (sh *Shoset) GetBindAddr() string {
	return sh.bindAddr
}

// GetName :
func (sh *Shoset) GetName() string {
	return sh.lName
}

// GetShosetType :
func (sh *Shoset) GetShosetType() string { return sh.ShosetType }

// String :
func (sh *Shoset) String() string {
	descr := fmt.Sprintf("Shoset { lName: %s,\nbindAddr: %s,\ntype: %s,\nbrothers %#v,\nNameBrothers %#v, joinConns\n%#v\n", sh.lName, sh.bindAddr, sh.ShosetType, sh.Brothers, sh.NameBrothers, sh.ConnsJoin)
	sh.ConnsByAddr.Iterate(
		func(addr string, conn *ShosetConn) {
			descr = fmt.Sprintf("%s - [%s] %s\n", descr, addr, conn.String())
		})
	descr += "}\n"
	// descr := fmt.Sprintf("Shoset { \nlName: %s,\nShosetType: %s,\nbinAddr: %s,\n", sh.lName, sh.ShosetType, sh.bindAddr)
	// descr += fmt.Sprintf("ConnsByAddr: %#v,\nConnsByName: %#v,\nConnsByType: %#v,\n", sh.ConnsByAddr, sh.ConnsByName, sh.ConnsByType)
	// descr += fmt.Sprintf("ConnsJoin %#v,\nConnsBye%#v,\nBrothers%#v,\nNameBrothers%#v\n}", sh.ConnsJoin, sh.ConnsJoin, sh.Brothers, sh.NameBrothers)
	return descr
}

//Link : Link to another Shoset
func (sh *Shoset) Link(address string) (*ShosetConn, error) {
	conn := new(ShosetConn)
	conn.ch = sh
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
func (sh *Shoset) Join(address string) (*ShosetConn, error) {

	exists := sh.ConnsJoin.Get(address)
	if exists != nil {
		return exists, nil
	}
	if address == sh.bindAddr {
		return nil, nil
	}

	conn := new(ShosetConn)
	conn.ch = sh
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

//SafeShutdown : safely disconnects a connection
func (sh *Shoset) SafeShutdown() error {

	// add all the connections (Join and Link) to the temporary list of connections
	// and remove said connection from the previous list
	fmt.Printf("____1____ConnsBye : \n%v\n", sh.ConnsBye.m)
	for connAddr, conn := range sh.ConnsByAddr.m {
		sh.ConnsBye.m[connAddr] = conn
		sh.deleteConn(connAddr)
	}
	fmt.Printf("____2____ConnsBye : \n%v\n", sh.ConnsBye.m)
	// for connAddr, conn := range sh.ConnsJoin.m {
	// 	//Apprently this delete the ConnsJoin too early
	// 	// and create pointers issues on shutdown....
	// 	sh.ConnsBye.m[connAddr] = conn
	// 	sh.ConnsJoin.Delete(connAddr)
	// }
	// fmt.Printf("____3____ConnsBye : \n%v\n", sh.ConnsBye.m)

	// use the temp list to send out one Bye msg to each connection
	cfgBye := msg.NewCfgBye(sh.bindAddr, sh.lName)
	for addr, conn := range sh.ConnsBye.m {
		if addr != sh.bindAddr {
			fmt.Printf("======> sending bye msg to %v  (%v)\n", addr, conn)
			errSend := conn.SendMessage(cfgBye)
			if errSend != nil {
				return errSend
			}
			go conn.runInConn()
		}
	}
	// wait for acknowledgements
	for len(sh.ConnsBye.m) > 0 {
		fmt.Printf("wait for list to empty : %v\n", sh.ConnsBye.m)
		time.Sleep(time.Second * time.Duration(5))
	}

	// then shutdown
	fmt.Printf("ready to be shutdown\n")
	//sh.socket.Close()

	return nil
}

func (sh *Shoset) deleteConn(connAddr string) {
	conn := sh.ConnsByAddr.Get(connAddr)
	if conn != nil {
		sh.ConnsByName.Delete(conn.name, connAddr)
		sh.ConnsByType.Delete(conn.ShosetType, connAddr)
		sh.ConnsByAddr.Delete(connAddr)
	}
}

// SetConn :
func (sh *Shoset) SetConn(connAddr, connType string, conn *ShosetConn) {
	if conn != nil {
		sh.ConnsByAddr.Set(connAddr, conn)
		sh.ConnsByType.Set(connType, conn.addr, conn)
		sh.ConnsByName.Set(conn.name, conn.addr, conn)
	}
}

//Bind : Connect to another Shoset
func (sh *Shoset) Bind(address string) error {
	if sh.bindAddr != "" {
		fmt.Println("Shoset already bound")
		return errors.New("Shoset already bound")
	}
	if sh.tlsServerOK == false {
		fmt.Println("TLS configuration not OK (certificate not found / loaded)")
		return errors.New("TLS configuration not OK (certificate not found / loaded)")
	}
	ipAddress, err := GetIP(address)
	if err != nil {
		return err
	}
	sh.bindAddr = ipAddress
	//fmt.Printf("Bind : handleBind adress %s", ipAddress)
	go sh.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (sh *Shoset) handleBind() error {
	listener, err := net.Listen("tcp", sh.bindAddr)
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
		tlsConn := tls.Server(unencConn, sh.tlsConfig)
		conn, _ := sh.inboudConn(tlsConn)
		//fmt.Printf("Shoset : accepted from %s", conn.addr)
		go conn.runInConn()
	}
	return nil
}

//inboudConn : Add a new connection from a client
func (sh *Shoset) inboudConn(tlsConn *tls.Conn) (*ShosetConn, error) {
	conn := new(ShosetConn)
	conn.socket = tlsConn
	conn.dir = "in"
	conn.ch = sh
	conn.addr = tlsConn.RemoteAddr().String()
	//sh.SetConn(conn.addr, conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	return conn, nil
}
