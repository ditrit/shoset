package shoset

import (
	"crypto/tls"
	"errors"
	"fmt"
	"sync"

	"net"
	"os"
	"strings"

	"github.com/ditrit/shoset/msg"
	uuid "github.com/kjk/betterguid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//terminal
// var certPath = "./certs/cert.pem"
// var keyPath = "./certs/key.pem"

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
	logger  zerolog.Logger
	Context map[string]interface{} //TOTO

	//	id          string
	ConnsByName        *MapSafeMapConn // map[string]map[string]*ShosetConn   connexions par nom logique
	LnamesByType       *MapSafeStrings // for gandalf
	LnamesByProtocol   *MapSafeStrings
	ConnsSingle        *MapSafeBool
	ConnsSingleAddress *SyncMapConn

	lName             string // Nom logique de la shoset
	ShosetType        string // Type logique de la shoset
	bindAddress       string // Adresse sur laquelle la shoset est bindée
	pkiRequestAddress string // La même que bindaddress mais seulement pour la pki cas la chaussette n'est pas encore bindée

	// Dictionnaire des queues de message (par type de message)
	Queue  map[string]*msg.Queue
	Get    map[string]func(*ShosetConn) (msg.Message, error)
	Handle map[string]func(*ShosetConn, msg.Message) error
	Send   map[string]func(*Shoset, msg.Message)
	Wait   map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message

	// configuration TLS
	tlsConfigSingleWay *tls.Config
	tlsConfigDoubleWay *tls.Config
	tlsServerOK        bool

	// synchronisation des goroutines
	Done chan bool

	viperConfig               *viper.Viper
	isValid                   bool
	isPki                     bool
	isCertified               bool
	listener                  net.Listener
	mu                        sync.RWMutex
	wentThroughPkiOnce        bool
	wentThroughHandleBindOnce bool
	fileName                  string
}

/*           Accessors            */
func (c *Shoset) GetBindAddress() string             { return c.bindAddress }
func (c *Shoset) GetPkiRequestAddress() string       { return c.pkiRequestAddress }
func (c *Shoset) GetLogicalName() string             { return c.lName }
func (c *Shoset) GetShosetType() string              { return c.ShosetType }
func (c *Shoset) GetIsValid() bool                   { return c.isValid }
func (c *Shoset) GetIsPki() bool                     { return c.isPki }
func (c *Shoset) GetIsCertified() bool               { return c.isCertified }
func (c *Shoset) GetWentThroughPkiOnce() bool        { return c.wentThroughPkiOnce }
func (c *Shoset) GetWentThroughHandleBindOnce() bool { return c.wentThroughHandleBindOnce }
func (c *Shoset) GetFileName() string                { return c.fileName }

func (c *Shoset) SetBindAddress(bindAddress string) {
	c.bindAddress = bindAddress
}

func (c *Shoset) SetPkiRequestAddress(pkiRequestAddress string) {
	c.pkiRequestAddress = pkiRequestAddress
}

func (c *Shoset) SetIsValid(state bool) {
	c.isValid = state
}
func (c *Shoset) SetIsPki(state bool) {
	c.isPki = state
}
func (c *Shoset) SetIsCertified(state bool) {
	c.isCertified = state
}

func (c *Shoset) SetWentThroughPkiOnce(state bool) {
	c.wentThroughPkiOnce = state
}

func (c *Shoset) SetWentThroughHandleBindOnce(state bool) {
	c.wentThroughHandleBindOnce = state
}

func (c *Shoset) SetFileName(fileName string) {
	c.fileName = fileName
}

/*       Constructor     */
func NewShoset(lName, ShosetType string) *Shoset { //l
	// Creation
	shst := Shoset{
		// Initialisation
		logger:             log.With().Str("uuid", uuid.New()).Logger(),
		Context:            make(map[string]interface{}),
		lName:              lName,
		ShosetType:         ShosetType,
		viperConfig:        viper.New(),
		ConnsByName:        NewMapSafeMapConn(),
		LnamesByType:       NewMapSafeStrings(),
		LnamesByProtocol:   NewMapSafeStrings(),
		isValid:            true,
		isPki:              false,
		isCertified:        false,
		listener:           nil,
		wentThroughPkiOnce: false,
		// isSingleTLS: true,
		ConnsSingle:        NewMapSafeBool(),
		ConnsSingleAddress: NewSyncMapConn(),

		// Dictionnaire des queues de message (par type de message)
		Queue:  make(map[string]*msg.Queue),
		Get:    make(map[string]func(*ShosetConn) (msg.Message, error)),
		Handle: make(map[string]func(*ShosetConn, msg.Message) error),
		Send:   make(map[string]func(*Shoset, msg.Message)),
		Wait:   make(map[string]func(*Shoset, *msg.Iterator, map[string]string, int) *msg.Message),

		// tlsConfig: &tls.Config{InsecureSkipVerify: true}
		tlsServerOK:        true,
		tlsConfigSingleWay: &tls.Config{InsecureSkipVerify: true},
		tlsConfigDoubleWay: nil,
	}

	shst.ConnsByName.SetViper(shst.viperConfig)

	shst.Queue["cfglink"] = msg.NewQueue()
	shst.Get["cfglink"] = GetConfigLink
	shst.Handle["cfglink"] = HandleConfigLink

	shst.Queue["cfgjoin"] = msg.NewQueue()
	shst.Get["cfgjoin"] = GetConfigJoin
	shst.Handle["cfgjoin"] = HandleConfigJoin

	shst.Queue["cfgbye"] = msg.NewQueue()
	shst.Get["cfgbye"] = GetConfigBye
	shst.Handle["cfgbye"] = HandleConfigBye

	shst.Queue["evt"] = msg.NewQueue()
	shst.Get["evt"] = GetEvent
	shst.Handle["evt"] = HandleEvent
	shst.Send["evt"] = SendEvent
	shst.Wait["evt"] = WaitEvent

	shst.Queue["pkievt"] = msg.NewQueue()
	shst.Get["pkievt"] = GetPkiEvent
	shst.Handle["pkievt"] = HandlePkiEvent
	shst.Send["pkievt"] = SendPkiEvent

	shst.Queue["cmd"] = msg.NewQueue()
	shst.Get["cmd"] = GetCommand
	shst.Handle["cmd"] = HandleCommand
	shst.Send["cmd"] = SendCommand
	shst.Wait["cmd"] = WaitCommand

	//TODO MOVE TO GANDALF
	shst.Queue["config"] = msg.NewQueue()
	shst.Get["config"] = GetConfig
	shst.Handle["config"] = HandleConfig
	shst.Send["config"] = SendConfig
	shst.Wait["config"] = WaitConfig

	shst.logger.Debug().Str("lname", lName).Msg("shoset created")
	return &shst
}

// Display with fmt - override the print of the object
func (c *Shoset) String() string {
	descr := fmt.Sprintf("Shoset -  lName: %s,\n\t\tbindAddr : %s,\n\t\ttype : %s, \n\t\tisPki : %t, \n\t\tisCertified : %t, \n\t\tConnsByName : ", c.GetLogicalName(), c.GetBindAddress(), c.GetShosetType(), c.GetIsPki(), c.GetIsCertified())
	for _, lName := range c.ConnsByName.Keys() {
		c.ConnsByName.Iterate(lName,
			func(key string, val *ShosetConn) {
				descr = fmt.Sprintf("%s %s\n\t\t\t     ", descr, val)
			})
	}
	// descr = fmt.Sprintf("%s \n\t\tLnamesByProtocol : MapSafeStrings{%s\n\t       ", descr, c.LnamesByProtocol)
	// descr = fmt.Sprintf("%s LnamesByType : MapSafeStrings{%s\n\t      ", descr, c.LnamesByType)
	return descr
}

//Bind : Connect to another Shoset
func (c *Shoset) Bind(address string) error {
	if c.GetBindAddress() != "" && c.GetBindAddress() != address { //socket already bounded to a port (already passed this Bind function once)
		return errors.New("Shoset already bound")
	}
	if !c.tlsServerOK { // TLS configuration not ok (security problem)
		return errors.New("TLS configuration not OK (certificate not found / loaded)")
	}

	// viper config
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.viperConfig.AddConfigPath(dirname + "/.shoset/" + c.GetFileName() + "/config/")
	c.viperConfig.SetConfigName(c.GetFileName())
	c.viperConfig.SetConfigType("yaml")

	if err := c.viperConfig.ReadInConfig(); err == nil {
		remotesToJoin, remotesToLink := c.ConnsByName.GetConfig() // get all the sockets we need to join
		for _, remote := range remotesToJoin {
			c.Protocol(address, remote, "join")
		}
		for _, remote := range remotesToLink {
			c.Protocol(address, remote, "link")
		}
	}

	if ipAddress, err := GetIP(address); err == nil {
		c.SetBindAddress(ipAddress) // bound to the port
	}

	//open a net listener
	if listener, err := net.Listen("tcp", c.GetBindAddress()); err != nil { // check if listener is ok
		return errors.New("a shoset is already listening on this port")
	} else {
		c.listener = listener
	}

	go c.handleBind()
	return nil
}

func (c *Shoset) handleBind() {
	defer c.listener.Close()

	for {
		// fmt.Println("waiting for new conn")
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			fmt.Println("error : Invalid connection for join - not the same type/name or shosetConn ended")
			return
		}
		unencConn, err := c.listener.Accept()
		if err != nil {
			fmt.Printf("serverShoset accept error: %s", err)
			break
		}
		// fmt.Println("accept conn")

		address_port := unencConn.RemoteAddr().String()
		address_parts := strings.Split(address_port, ":")
		address_ := address_parts[0]

		if c.ConnsSingle.Get(address_) {
			// fmt.Println(c.GetBindAddress(), "trying singleWay")
			tlsConn := tls.Server(unencConn, c.tlsConfigSingleWay) // create the securised connection protocol

			address := tlsConn.RemoteAddr().String()
			conn, _ := NewShosetConn(c, address, "in") // create the securised connection
			conn.socket = tlsConn                      //we override socket attribut with our securised protocol

			// fmt.Println(c.GetBindAddress(), "enters single")
			go conn.runInConnSingle(address_)
		} else {
			// fmt.Println(c.GetBindAddress(), "trying doubleWay")
			tlsConn := tls.Server(unencConn, c.tlsConfigDoubleWay) // create the securised connection protocol

			address := tlsConn.RemoteAddr().String()
			conn, _ := NewShosetConn(c, address, "in") // create the securised connection
			conn.socket = tlsConn                      //we override socket attribut with our securised protocol

			// fmt.Println("sending msg")
			_, err = conn.socket.Write([]byte("hello double\n"))
			// fmt.Println("msg sent")
			if err == nil {
				// fmt.Println(c.GetBindAddress(), "enters double and exit single")
				// if !c.GetWentThroughHandleBindOnce() {
				// 	c.SetWentThroughHandleBindOnce(true)
				// 	go conn.runInConnDouble()
				// }
				go conn.runInConnDouble()
				// return nil
			} else {
				// fmt.Println("err double : ", err)
				c.ConnsSingle.Set(address_, true)
				conn.socket.Close()
			}
		}
	}
}

func (c *Shoset) Protocol(bindAddress, remoteAddress, protocolType string) {
	// init cert if needed
	c.logger.Debug().Strs("params", []string{bindAddress, remoteAddress, protocolType}).Msg("protocol init")
	if !c.GetIsCertified() && !c.GetWentThroughPkiOnce() {
		ipAddress, err := GetIP(bindAddress) // parse the address from function parameter to get the IP
		if err != nil {
			// IP nok -> return early
			fmt.Println("wrong IP format:", err)
			return
		}

		_ipAddress := strings.Replace(ipAddress, ":", "_", -1)
		_ipAddress = strings.Replace(_ipAddress, ".", "-", -1)

		_, err = InitConfFolder(_ipAddress)
		if err != nil { // initialization of folder did'nt work
			fmt.Println(err)
			return
		}

		// set filename _after_ successful conf creation
		c.SetFileName(_ipAddress)
		c.SetPkiRequestAddress(ipAddress)
		initConn, err := NewShosetConn(c, remoteAddress, "out")
		if err != nil {
			fmt.Println("couldn't create shoset:", err)
			return
		}
		err = initConn.runPkiRequest() // I don't have my certs, I request them
		if err != nil {
			fmt.Println(c.GetPkiRequestAddress(), "runPkiRequest didn't work", err)
			return
		}
		if !c.GetIsCertified() {
			fmt.Println("couldn't certify")
			return
		}
		initConn.socket.Close()
		c.SetWentThroughPkiOnce(true) // avoid concurrency when multiple protocols are running at the same time
	}

	if c.GetBindAddress() == "" {
		err := c.Bind(bindAddress) // I have my certs, I can bind
		if err != nil {
			fmt.Println("Couldn't bind")
		}
	}

	conn, _ := NewShosetConn(c, remoteAddress, "out")
	switch protocolType {
	case "join":
		conns := c.ConnsByName.Get(c.GetLogicalName())
		if conns != nil {
			exists := conns.Get(remoteAddress) // check if remoteAddress is already in the map
			if exists != nil {                 //connection already established for this socket
				return
			}
		}
		if remoteAddress == c.GetBindAddress() { // connection impossible with itself
			return
		}
		go conn.runJoinConn()
	case "link":
		conns := c.ConnsByName.Get(c.GetLogicalName())
		if conns != nil {
			exists := conns.Get(remoteAddress) // check if remoteAddress is already in the map
			if exists != nil {                 //connection already established for this socket
				return
			}
		}
		if remoteAddress == c.GetBindAddress() { // connection impossible with itself
			return
		}
		go conn.runLinkConn()
	case "bye":
		go conn.runByeConn()
	default:
		return
	}
}

func (c *Shoset) deleteConn(connAddr, connLname string) {
	// fmt.Println(c.GetBindAddress(), " enter deleteConn")
	if conns := c.ConnsByName.Get(connLname); conns != nil {
		if conns.Get(connAddr) != nil {
			// fmt.Println(c.GetBindAddress(), " is ok in deleteConn")
			c.ConnsByName.Delete(connLname, connAddr, c.fileName)
		}
	}
}

func (c *Shoset) GetConnsByTypeArray(shosetType string) []*ShosetConn {
	lNames := c.LnamesByType.Keys(shosetType)
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
