package shoset

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/ditrit/shoset/msg"
)

//terminal
var certPath = "./certs/cert.pem"
var keyPath = "./certs/key.pem"

//debugger
// var certPath = "../certs/cert.pem"
// var keyPath = "../certs/key.pem"

// returns bool whether the given file or directory exists
func CertsCheck(path string) bool {
	_, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return false
	}
	return true
}

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
	Queue  map[string]*msg.Queue
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

/*           Accessors            */
func (c Shoset) GetBindAddr() string   { return c.bindAddr }
func (c Shoset) GetName() string       { return c.lName }
func (c Shoset) GetShosetType() string { return c.ShosetType }

/*       Constructor     */
func NewShoset(lName, ShosetType string) *Shoset { //l
	// Creation
	shoset := Shoset{}

	// Initialisation
	shoset.lName = lName
	shoset.ShosetType = ShosetType
	shoset.ConnsByAddr = NewMapSafeConn()
	shoset.ConnsByName = NewMapSafeMapConn()
	shoset.ConnsByType = NewMapSafeMapConn()
	shoset.ConnsJoin = NewMapSafeConn()
	shoset.Brothers = NewMapSafeBool()
	shoset.NameBrothers = NewMapSafeBool()

	// Dictionnaire des queues de message (par type de message)
	shoset.Queue = make(map[string]*msg.Queue)
	shoset.Get = make(map[string]func(*ShosetConn) (msg.Message, error))
	shoset.Handle = make(map[string]func(*ShosetConn, msg.Message) error)
	shoset.Send = make(map[string]func(*Shoset, msg.Message))
	shoset.Wait = make(map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message)

	shoset.Queue["cfglink"] = msg.NewQueue()
	shoset.Get["cfglink"] = GetConfigLink
	shoset.Handle["cfglink"] = HandleConfigLink

	shoset.Queue["cfgjoin"] = msg.NewQueue()
	shoset.Get["cfgjoin"] = GetConfigJoin
	shoset.Handle["cfgjoin"] = HandleConfigJoin

	shoset.Queue["evt"] = msg.NewQueue()
	shoset.Get["evt"] = GetEvent
	shoset.Handle["evt"] = HandleEvent
	shoset.Send["evt"] = SendEvent
	shoset.Wait["evt"] = WaitEvent

	shoset.Queue["cmd"] = msg.NewQueue()
	shoset.Get["cmd"] = GetCommand
	shoset.Handle["cmd"] = HandleCommand
	shoset.Send["cmd"] = SendCommand
	shoset.Wait["cmd"] = WaitCommand

	//TODO MOVE TO GANDALF
	shoset.Queue["config"] = msg.NewQueue()
	shoset.Get["config"] = GetConfig
	shoset.Handle["config"] = HandleConfig
	shoset.Send["config"] = SendConfig
	shoset.Wait["config"] = WaitConfig

	// Configuration TLS //////////////////////////////////// à améliorer
	if CertsCheck(certPath) && CertsCheck(keyPath) {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil { // only client in insecure mode
			fmt.Println("! Unable to Load certificate !")
			shoset.tlsConfig = &tls.Config{InsecureSkipVerify: true}
			shoset.tlsServerOK = false
		} else {
			shoset.tlsConfig = &tls.Config{
				Certificates:       []tls.Certificate{cert},
				InsecureSkipVerify: true,
			}
			shoset.tlsServerOK = true
		}
	} else {
		fmt.Println("! Unable to Load certificate !")
		shoset.tlsServerOK = false
	}
	return &shoset
}

// Display with fmt - override the print of the object
func (c Shoset) String() string {
	descr := fmt.Sprintf("Shoset - lName: %s,\n\t\tbindAddr : %s,\n\t\ttype : %s,\n\t\tjoinConns : %#v\n", c.lName, c.bindAddr, c.ShosetType, c.ConnsJoin)
	c.ConnsByAddr.Iterate(
		func(key string, val *ShosetConn) {
			descr = fmt.Sprintf("%s - [%s] %s\n", descr, key, val.String())
		})
	return descr
}

//Link : Link to another Shoset
func (c *Shoset) Link(address string) (*ShosetConn, error) {
	conn, _ := NewShosetConn(c, address, "out")
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
	conn, _ := NewShosetConn(c, address, "out")
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
	if c.bindAddr != "" { //socket already bounded to a port (already passed this Bind function once)
		fmt.Println("Shoset already bound")
		return errors.New("Shoset already bound")
	}
	if !c.tlsServerOK { // TLS configuration not ok (security problem)
		fmt.Println("TLS configuration not OK (certificate not found / loaded)")
		return errors.New("TLS configuration not OK (certificate not found / loaded)")
	}
	ipAddress, err := GetIP(address) // parse the address from function parameter to get the IP
	if err != nil {                  // check if IP is ok
		return err
	}
	c.bindAddr = ipAddress // bound to the port
	go c.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (c *Shoset) handleBind() error {
	listener, err := net.Listen("tcp", c.bindAddr) //open a net listener
	if err != nil {                                // check if listener is ok
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
		tlsConn := tls.Server(unencConn, c.tlsConfig) // create the securised connection
		conn, _ := c.inBoundConn(tlsConn)
		go conn.runInConn()
	}
	return nil
}

//inBoundConn : Add a new connection from a client
func (c *Shoset) inBoundConn(tlsConn *tls.Conn) (*ShosetConn, error) {
	address := tlsConn.RemoteAddr().String()
	//c.SetConn(conn.addr, conn)
	conn, _ := NewShosetConn(c, address, "in")
	conn.socket = tlsConn //we override socket attribut with our new value
	return conn, nil
}
