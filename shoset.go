package shoset

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/ditrit/shoset/msg"
	uuid "github.com/kjk/betterguid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// MessageHandlers interface
type MessageHandlers interface {
	Get(c *ShosetConn) (msg.Message, error)
	HandleDoubleWay(c *ShosetConn, message msg.Message) error
	Send(c *Shoset, m msg.Message)
	Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message
}

//Shoset :
type Shoset struct {
	logger  zerolog.Logger
	Context map[string]interface{} // used for gandalf

	ConnsByName      *MapSyncMap // map[lName]map[remoteAddress]*ShosetConn   connexions par nom logique
	LnamesByType     *MapSyncMap // map[shosetType]map[lName]bool used for gandalf
	LnamesByProtocol *MapSyncMap // map[protocolType]map[lName]bool
	ConnsSingleBool  *sync.Map   // map[ipAddress]bool ipAddresses waiting in singleWay to be handled for tls double way
	ConnsSingleConn  *sync.Map   // map[ipAddress]*ShosetConn ShosetConns waiting in singleWay to be handled for tls double way

	lName       string // Nom logique de la shoset
	ShosetType  string // Type logique de la shoset
	bindAddress string // Adresse sur laquelle la shoset est bindée
	config      *Config
	isValid     bool
	isPki       bool
	listener    net.Listener

	// configuration TLS
	tlsConfigSingleWay *tls.Config
	tlsConfigDoubleWay *tls.Config

	// Dictionnaire des queues de message (par type de message)
	Queue    map[string]*msg.Queue
	Handlers map[string]MessageHandlers

	// synchronisation des goroutines
	Done chan bool
}

func (c *Shoset) GetBindAddress() string { return c.bindAddress }
func (c *Shoset) GetLogicalName() string { return c.lName }
func (c *Shoset) GetShosetType() string  { return c.ShosetType }
func (c *Shoset) GetIsValid() bool       { return c.isValid }
func (c *Shoset) GetIsPki() bool         { return c.isPki }

func (c *Shoset) GetConnsByTypeArray(shosetType string) []*ShosetConn {
	lNames := c.LnamesByType.Keys(shosetType)
	var connsByType []*ShosetConn
	for _, lName := range lNames {
		lNameMap, _ := c.ConnsByName.smap.Load(lName)
		keys := Keys(lNameMap.(*sync.Map), ALL)
		for _, key := range keys {
			conn, _ := lNameMap.(*sync.Map).Load(key)
			connsByType = append(connsByType, conn.(*ShosetConn))
		}
	}
	return connsByType
}

func (c *Shoset) IsCertified(path string) bool {
	if fileExists(path + "privateCAKey.key") {
		c.SetIsPki(true)
	}
	return fileExists(path)
}

func (c *Shoset) SetBindAddress(bindAddress string) { c.bindAddress = bindAddress }
func (c *Shoset) SetIsValid(state bool)             { c.isValid = state }
func (c *Shoset) SetIsPki(state bool)               { c.isPki = state }

func NewShoset(lName, ShosetType string) *Shoset {
	shst := Shoset{
		logger:           log.With().Str("uuid", uuid.New()).Logger(),
		Context:          make(map[string]interface{}),
		lName:            lName,
		ShosetType:       ShosetType,
		config:           NewConfig(),
		ConnsByName:      new(MapSyncMap),
		LnamesByType:     new(MapSyncMap),
		LnamesByProtocol: new(MapSyncMap),
		isValid:          true,
		isPki:            false,
		listener:         nil,
		ConnsSingleBool:  new(sync.Map),
		ConnsSingleConn:  new(sync.Map),

		// Dictionnaire des queues de message (par type de message)
		Queue:    make(map[string]*msg.Queue),
		Handlers: make(map[string]MessageHandlers),

		tlsConfigSingleWay: &tls.Config{InsecureSkipVerify: true},
		tlsConfigDoubleWay: nil,
	}

	shst.ConnsByName.SetConfig(shst.config)

	shst.Queue["cfglink"] = msg.NewQueue()
	shst.Handlers["cfglink"] = new(ConfigLinkHandler)

	shst.Queue["cfgjoin"] = msg.NewQueue()
	shst.Handlers["cfgjoin"] = new(ConfigJoinHandler)

	shst.Queue["cfgbye"] = msg.NewQueue()
	shst.Handlers["cfgbye"] = new(ConfigByeHandler)

	shst.Queue["evt"] = msg.NewQueue()
	shst.Handlers["evt"] = new(EventHandler)

	shst.Queue["pkievt_TLSdoubleWay"] = msg.NewQueue()
	shst.Handlers["pkievt_TLSdoubleWay"] = new(PkiEventHandler)

	shst.Queue["pkievt_TLSsingleWay"] = msg.NewQueue()
	shst.Handlers["pkievt_TLSsingleWay"] = new(PkiEventHandler)

	shst.Queue["cmd"] = msg.NewQueue()
	shst.Handlers["cmd"] = new(CommandHandler)

	//TODO MOVE TO GANDALF
	shst.Queue["config"] = msg.NewQueue()
	shst.Handlers["config"] = new(ConfigHandler)

	shst.logger.Debug().Str("lname", lName).Msg("shoset created")
	return &shst
}

// String returns the formatted string of Shoset object.
func (c *Shoset) String() string {
	return c.PrettyPrint()
}

// PrettyPrint returns the indented string of Shoset object.
func (c *Shoset) PrettyPrint() string {
	descr := fmt.Sprintf("Shoset{\n\t- lName: %s,\n\t- bindAddr : %s,\n\t- type : %s, \n\t- isPki : %t, \n\t- ConnsByName:", c.GetLogicalName(), c.GetBindAddress(), c.GetShosetType(), c.GetIsPki())

	c.ConnsByName.Iterate(
		func(key string, val interface{}) {
			descr += fmt.Sprintf("\n\t\t* %s", val)
		})

	descr += "\n}\n"
	return descr
}

//Bind : Connect to another Shoset
func (c *Shoset) Bind(address string) error {
	if err := c.config.ReadConfig(c.config.GetFileName()); err == nil {
		for _, remote := range c.ConnsByName.cfg.GetSlice(PROTOCOL_JOIN) {
			c.Protocol(address, remote, PROTOCOL_JOIN)
		}
		for _, remote := range c.ConnsByName.cfg.GetSlice(PROTOCOL_LINK) {
			c.Protocol(address, remote, PROTOCOL_LINK)
		}
	}

	ipAddress, err := GetIP(address)
	if err != nil {
		c.logger.Error().Msg("couldn't set bindAddress : " + err.Error())
		return err
	}
	c.SetBindAddress(ipAddress) // bound to the port

	listener, err := net.Listen(CONNECTION_TYPE, DEFAULT_IP+strings.Split(ipAddress, ":")[1]) // listen on each ipaddresses
	if err != nil {
		c.logger.Error().Msg("listen error : " + err.Error())
		return err
	}
	c.listener = listener

	go c.handleBind()
	return nil
}

func (c *Shoset) handleBind() {
	defer c.listener.Close()

	for {
		unencConn, err := c.listener.Accept()
		if err != nil {
			c.logger.Error().Msg("serverShoset accept error : " + err.Error())
			return
		}

		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			c.logger.Error().Msg("Invalid connection for join - not the same type/name or shosetConn ended")
			return
		}

		if exists, _ := c.ConnsSingleBool.Load(strings.Split(unencConn.RemoteAddr().String(), ":")[0]); exists != nil { // get ipAddress
			tlsConn := tls.Server(unencConn, c.tlsConfigSingleWay) // create the securised connection protocol

			conn, _ := NewShosetConn(c, unencConn.RemoteAddr().String(), IN) // create the securised connection
			conn.UpdateConn(tlsConn)                                         //we override socket attribut with our securised protocol

			go conn.runInConnSingle(strings.Split(unencConn.RemoteAddr().String(), ":")[0])
		} else {
			tlsConn := tls.Server(unencConn, c.tlsConfigDoubleWay)
			conn, _ := NewShosetConn(c, unencConn.RemoteAddr().String(), IN)
			conn.UpdateConn(tlsConn)

			_, err = conn.socket.Write([]byte(TLS_DOUBLE_WAY_TEST_WRITE + "\n"))
			if err == nil {
				go conn.runInConnDouble()
			} else {
				c.ConnsSingleBool.Store(strings.Split(unencConn.RemoteAddr().String(), ":")[0], true) // set ipAddress
			}
		}
	}
}

func (c *Shoset) Protocol(bindAddress, remoteAddress, protocolType string) {
	c.logger.Debug().Strs("params", []string{bindAddress, remoteAddress, protocolType}).Msg("protocol init")
	ipAddress, err := GetIP(bindAddress) // parse the address from function parameter to get the IP
	if err != nil {
		// IP nok -> return early
		c.logger.Error().Msg("wrong IP format : " + err.Error())
		return
	}
	formatedIpAddress := strings.Replace(ipAddress, ":", "_", -1)
	formatedIpAddress = strings.Replace(formatedIpAddress, ".", "-", -1)

	// init cert if needed
	if !c.IsCertified(c.config.baseDir + formatedIpAddress + PATH_CERT_DIR) {
		c.logger.Debug().Msg("ask certification")

		_, err = c.config.InitFolders(formatedIpAddress)
		if err != nil { // initialization of folders did not work
			c.logger.Error().Msg("couldn't init folder: " + err.Error())
			return
		}

		// set filename _after_ successful conf creation
		c.config.SetFileName(formatedIpAddress)

		err = c.Certify(bindAddress, remoteAddress)
		if err != nil {
			return
		}
	}

	if c.GetBindAddress() == VOID {
		err := c.Bind(bindAddress) // I have my certs, I can bind
		if err != nil {
			c.logger.Error().Msg("couldn't set bindAddress : " + err.Error())
			return
		}
	}

	if remoteAddress == c.GetBindAddress() { // connection impossible with itself
		c.logger.Error().Msg("can't protocol on itself")
		return
	}

	if conns, _ := c.ConnsByName.smap.Load(c.GetLogicalName()); conns != nil {
		if exists, _ := conns.(*sync.Map).Load(remoteAddress); exists != nil { // check if remoteAddress is already in the map
			c.logger.Warn().Msg("already did a protocol on this shoset")
			return
		}
	}

	conn, _ := NewShosetConn(c, remoteAddress, OUT)
	switch protocolType {
	case PROTOCOL_JOIN:
		go conn.runJoinConn()
	case PROTOCOL_LINK:
		go conn.runLinkConn()
	case PROTOCOL_EXIT:
		conn.SetRemoteAddress(conn.GetLocalAddress()) // needed otherwise remoteAddress will not be considered for the bye
		go conn.runByeConn()
	default:
		return
	}
}

func (c *Shoset) deleteConn(connAddr, connLname string) {
	if conns, _ := c.ConnsByName.smap.Load(connLname); conns != nil {
		if conn, _ := conns.(*sync.Map).Load(connAddr); conn != nil {
			c.ConnsByName.Delete(connLname, connAddr, c.config.GetFileName())
		}
	}
}

func (c *Shoset) Send(msg msg.Message) { //Use pointer for msg ?
	fmt.Println("msg (Send)",msg)
	c.Handlers[msg.GetMsgType()].Send(c, msg)
}

func (c *Shoset) Wait(msgType string, args map[string]string, timeout int) *msg.Message {
	iter := msg.NewIterator(c.Queue[msgType])
	fmt.Println("Iterator : ", iter)
	return c.Handlers[msgType].Wait(c, iter, args, timeout)
}

// func (c *Shoset) newIter(msgType string) {
//     msg.NewIterator(*msg.Queue[msgType])
// }
