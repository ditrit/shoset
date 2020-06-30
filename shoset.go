package shoset

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
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
	certs       map[string][]byte
	tlsConfig   *tls.Config
	tlsServerOK bool
	canSignCert bool

	// synchronisation des goroutines
	Done chan bool
}

// NewShoset : constructor
func NewShoset(lName, ShosetType string, certMap map[string][]byte) *Shoset {
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

	c.Queue["csr"] = msg.NewQueue()
	c.Get["csr"] = GetCertRequest
	c.Handle["csr"] = HandleCertRequest

	//TODO MOVE TO GANDALF
	c.Queue["config"] = msg.NewQueue()
	c.Get["config"] = GetConfig
	c.Handle["config"] = HandleConfig
	c.Send["config"] = SendConfig
	c.Wait["config"] = WaitConfig

	// Configuration TLS
	c.certs = certMap
	c.loadTLSConfig()

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

func (c *Shoset) loadTLSConfig() {
	// Check CA
	caCertPool := x509.NewCertPool()
	ok := checkCA(c.certs["ca"]) && caCertPool.AppendCertsFromPEM(c.certs["ca"])
	if !ok {
		log.Fatalf("shoset error: valid root (ca) certificate required\n")
	}
	// Check CA private key: if given that means the shoset is allowed to sign
	if c.certs["cakey"] != nil {
		if _, err := tls.X509KeyPair(c.certs["ca"], c.certs["cakey"]); err != nil {
			log.Fatalf("shoset error: invalid ca key\n")
		}
		c.canSignCert = true
	}
	// Default conf : no client check && no server
	c.tlsConfig = &tls.Config{
		RootCAs:   caCertPool,
		ClientCAs: caCertPool,
	}
	c.tlsServerOK = true
	// If there is no cert and/or key
	if c.certs["key"] == nil || c.certs["cert"] == nil {
		key, pub, err := genPrivKey()
		if err != nil {
			log.Fatalf("shoset error: error while generating private key : %v\n", err)
		}
		c.certs["key"] = key
		c.certs["pubkey"] = pub

		return
	}
	// Loading cert-key pair
	cert, err := tls.X509KeyPair(c.certs["cert"], c.certs["key"])
	if err != nil || !ok { // only client in insecure mode
		fmt.Println("Unable to Load certificate")
		return
	}
	c.tlsConfig.Certificates = []tls.Certificate{cert}
	c.tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	c.tlsServerOK = true
	return
}

func (c *Shoset) reloadTLSConfig() {
	c.tlsConfig = &tls.Config{
		RootCAs:   c.tlsConfig.RootCAs,
		ClientCAs: c.tlsConfig.ClientCAs,
	}
	c.tlsServerOK = true
	cert, err := tls.X509KeyPair(c.certs["cert"], c.certs["key"])
	if err != nil && !c.canSignCert {
		return
	}
	if err != nil {
		certPEM, err := signCert(strings.Split(c.bindAddr, ":")[0], c.certs["pub"], c.certs["ca"], c.certs["cakey"])
		if err != nil {
			return
		}
		cert, err = tls.X509KeyPair(certPEM, c.certs["key"])
		if err != nil {
			return
		}
	}
	c.tlsConfig.Certificates = []tls.Certificate{cert}
	c.tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	c.tlsServerOK = true
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
	if c.tlsServerOK {
		go conn.runOutConn(conn.addr)
	} else {
		go conn.runOutCSROnly(conn.addr)
	}
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

// DeleteConn :
func (c *Shoset) DeleteConn(connAddr string) {
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
	c.reloadTLSConfig()
	//fmt.Printf("Bind : handleBind adress %s", ipAddress)
	for {
		if c.tlsServerOK {
			fmt.Printf("%s bind starting\n", c.GetName())
			break
		}
		time.Sleep(time.Millisecond * time.Duration(50))
	}
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
			fmt.Printf("serverShoset accept error: %s\n", err)
			break
		}
		tlsConn := tls.Server(unencConn, c.tlsConfig)
		conn, _ := c.inboudConn(tlsConn)
		//fmt.Printf("Shoset : accepted from %s\n", conn.addr)
		if len(conn.socket.ConnectionState().PeerCertificates) > 0 {
			go conn.runInCSROnly()
		} else {
			go conn.runInConn()
		}
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
