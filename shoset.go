package shoset

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ditrit/shoset/msg"
	"github.com/spf13/viper"
)

//terminal
var certPath = "./certs/cert.pem"
var keyPath = "./certs/key.pem"

//debugger
// var certPath = "../certs/cert.pem"
// var keyPath = "../certs/key.pem"

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
	ConnsByName      *MapSafeMapConn // map[string]map[string]*ShosetConn   connexions par nom logique
	LnamesByType     *MapSafeStrings // for gandalf
	LnamesByProtocol *MapSafeStrings

	lName       string // Nom logique de la shoset
	ShosetType  string // Type logique de la shoset
	bindAddress string // Adresse sur laquelle la shoset est bindée

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

	viperConfig *viper.Viper
	isValid     bool
}

/*           Accessors            */
func (c Shoset) GetBindAddress() string { return c.bindAddress }
func (c Shoset) GetLogicalName() string { return c.lName }
func (c Shoset) GetShosetType() string  { return c.ShosetType }
func (c *Shoset) GetIsValid() bool      { return c.isValid }

func (c *Shoset) SetBindAddress(bindAddress string) {
	if bindAddress != "" {
		c.bindAddress = bindAddress
	}
}

func (c *Shoset) SetIsValid(state bool) {
	c.isValid = state
}

/*       Constructor     */
func NewShoset(lName, ShosetType string) *Shoset { //l
	// Creation
	shoset := Shoset{}

	// Initialisation
	shoset.Context = make(map[string]interface{})

	shoset.lName = lName
	shoset.ShosetType = ShosetType
	shoset.viperConfig = viper.New()
	shoset.ConnsByName = NewMapSafeMapConn()
	shoset.LnamesByType = NewMapSafeStrings()
	shoset.LnamesByProtocol = NewMapSafeStrings()
	shoset.ConnsByName.SetViper(shoset.viperConfig)
	shoset.isValid = true

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

	shoset.Queue["cfgbye"] = msg.NewQueue()
	shoset.Get["cfgbye"] = GetConfigBye
	shoset.Handle["cfgbye"] = HandleConfigBye

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
	if pathCheck(certPath) && pathCheck(keyPath) {
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
	descr := fmt.Sprintf("Shoset -  lName: %s,\n\t\tbindAddr : %s,\n\t\ttype : %s, \n\t\tConnsByName : ", c.GetLogicalName(), c.GetBindAddress(), c.GetShosetType())
	for _, lName := range c.ConnsByName.Keys() {
		c.ConnsByName.Iterate(lName,
			func(key string, val *ShosetConn) {
				descr = fmt.Sprintf("%s %s\n\t\t\t     ", descr, val)
			})
	}
	descr = fmt.Sprintf("%s \n\t\tLnamesByProtocol : MapSafeStrings{%s\n\t       ", descr, c.LnamesByProtocol)
	descr = fmt.Sprintf("%s LnamesByType : MapSafeStrings{%s\n\t      ", descr, c.LnamesByType)
	// c.LnamesByType.Iterate(
	// 	func(key string, val map[string]bool) {
	// 		descr = fmt.Sprintf("%s %s\n\t\t\t     ", descr, val)
	// 	})
	// c.LnamesByProtocol.Iterate(
	// 	func(key string, val map[string]bool) {
	// 		descr = fmt.Sprintf("%s %s\n\t\t\t     ", descr, val)
	// 	})
	return descr
}

//Bind : Connect to another Shoset
func (c *Shoset) Bind(address string) error {
	if c.GetBindAddress() != "" { //socket already bounded to a port (already passed this Bind function once)
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

	viperAddress := computeAddress(ipAddress)
	c.ConnsByName.SetConfigName(viperAddress)

	// viper config
	dirname, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	if !pathCheck(dirname + "/config/") {
		os.Mkdir(dirname+"/config/", 0777)
	}

	c.viperConfig.AddConfigPath(dirname+"/config/")
	c.viperConfig.SetConfigName(viperAddress)
	c.viperConfig.SetConfigType("yaml")

	if err := c.viperConfig.ReadInConfig(); err != nil {
	} else {
		remotesToJoin, remotesToLink := c.ConnsByName.GetConfig() // get all the sockets we need to join
		for _, remote := range remotesToJoin {
			c.Protocol(remote, "join")
		}
		for _, remote := range remotesToLink {
			c.Protocol(remote, "link")
		}
	}

	c.SetBindAddress(ipAddress) // bound to the port
	go c.handleBind()           // process runInconn()
	return nil
}

// runBindTo : handler for the socket
func (c *Shoset) handleBind() error {
	listener, err := net.Listen("tcp", c.GetBindAddress()) //open a net listener
	if err != nil {                                        // check if listener is ok
		fmt.Println("Failed to bind:", err.Error())
		return err
	}
	// defer WriteViper()
	defer listener.Close()

	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}
		unencConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("serverShoset accept error: %s", err)
			break
		}
		tlsConn := tls.Server(unencConn, c.tlsConfig) // create the securised connection protocol
		address := tlsConn.RemoteAddr().String()
		conn, _ := NewShosetConn(c, address, "in") // create the securised connection
		conn.socket = tlsConn                      //we override socket attribut with our securised protocol
		go conn.runInConn()
	}
	return nil
}

func (c *Shoset) Protocol(address, protocolType string) (*ShosetConn, error) {
	var conn *ShosetConn
	switch protocolType {
	case "join":
		conns := c.ConnsByName.Get(c.GetLogicalName())
		if conns != nil {
			exists := conns.Get(address) // check if address is already in the map
			if exists != nil {           //connection already established for this socket
				return exists, nil
			}
		}
		if address == c.GetBindAddress() { // connection impossible with itself
			return nil, nil
		}
		conn, _ := NewShosetConn(c, address, "out")
		go conn.runJoinConn()
	case "link":
		conns := c.ConnsByName.Get(c.GetLogicalName())
		if conns != nil {
			exists := conns.Get(address) // check if address is already in the map
			if exists != nil {           //connection already established for this socket
				return exists, nil
			}
		}
		if address == c.GetBindAddress() { // connection impossible with itself
			return nil, nil
		}
		conn, _ := NewShosetConn(c, address, "out")
		go conn.runOutConn()
	case "bye":
		conn, _ := NewShosetConn(c, address, "out")
		go conn.runEndConn()
	default:
		fmt.Println("Wrong input protocolType")
		return nil, errors.New("wrong input protocolType")
	}
	return conn, nil
}

func (c *Shoset) deleteConn(connAddr, connLname string) {
	// fmt.Println(c.GetBindAddress(), " enter deleteConn")
	if conns := c.ConnsByName.Get(connLname); conns != nil {
		if conns.Get(connAddr) != nil {
			// fmt.Println(c.GetBindAddress(), " is ok in deleteConn")
			c.ConnsByName.Delete(connLname, connAddr)
		}
	}
}

// compute the name for the .yaml file corresponding to each socket
func computeAddress(ipAddress string) string {
	_ipAddress := strings.Replace(ipAddress, ":", "_", 1)
	_ipAddress = strings.Replace(_ipAddress, ".", "~", 3)
	name := "shoset_" + _ipAddress
	return name
}

// returns bool whether the given file or directory exists
func pathCheck(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		fmt.Println("File does not exist.")
		return false
	}
	return true
}

func (c *Shoset) GetConnsByType(shosetType string) map[string]*ShosetConn {
	lNames := c.LnamesByType.Keys(shosetType)
	connsByType := make(map[string]*ShosetConn)
	for _, lName := range lNames {
		lNameMap := c.ConnsByName.Get(lName)
		keys := lNameMap.Keys("all")
		for _, key := range keys {
			connsByType[key] = lNameMap.Get(key)
		}
	}
	return connsByType
}

func (c *Shoset) GetConnsByTypeArray(shosetType string) []*ShosetConn {
	lNames := c.LnamesByType.Keys(shosetType)
	// fmt.Println("lNames : ", lNames)
	var connsByType []*ShosetConn
	for _, lName := range lNames {
		lNameMap := c.ConnsByName.Get(lName)
		keys := lNameMap.Keys("all")
		for _, key := range keys {
			connsByType = append(connsByType, lNameMap.Get(key))
		}
	}
	return connsByType
}
