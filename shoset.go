package shoset

import (
	"crypto/tls"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"

	"github.com/ditrit/shoset/msg"
	uuid "github.com/kjk/betterguid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// MessageHandlers interface.
// Each handler must implement this interface
type MessageHandlers interface {
	Get(c *ShosetConn) (msg.Message, error)
	HandleDoubleWay(c *ShosetConn, message msg.Message) error
	Send(s *Shoset, m msg.Message)
	Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message
}

//Shoset : smart object based on network socket but with upgraded features
type Shoset struct {
	Logger zerolog.Logger // pretty logger

	Context map[string]interface{} // used for gandalf

	ConnsByLname     *MapSyncMap // map[lName]map[remoteAddress]*ShosetConn   connections by logical name
	LnamesByType     *MapSyncMap // map[shosetType]map[lName]bool used for gandalf
	LnamesByProtocol *MapSyncMap // map[protocolType]map[lName]bool
	ConnsSingleBool  *sync.Map   // map[ipAddress]bool ipAddresses waiting in singleWay to be handled for TLS double way
	ConnsSingleConn  *sync.Map   // map[ipAddress]*ShosetConn ShosetConns waiting in singleWay to be handled for TLS double way

	RouteTable *sync.Map //map[string] *Router

	bindAddress string       // address on which is bound the Shoset
	logicalName string       // logical name of the Shoset
	shosetType  string       // logical type of the shoset
	isValid     bool         // state of the Shoset - must be done differently in future review
	isPki       bool         // is the Shoset admin of network or not
	listener    net.Listener // generic network listener for stream-oriented protocols

	tlsConfigSingleWay *tls.Config // mutual authentication between client and server
	tlsConfigDoubleWay *tls.Config // client authenticate server

	Queue    map[string]*msg.Queue      // map for message queueing
	Handlers map[string]MessageHandlers // map for message handling

	Done chan bool // goroutines synchronization
}

// GetBindAddress returns bindAddress from Shoset.
func (s *Shoset) GetBindAddress() string { return s.bindAddress }

// GetLogicalName returns lName from Shoset.
func (s *Shoset) GetLogicalName() string { return s.logicalName }

// GetShosetType returns shosetType from Shoset.
func (s *Shoset) GetShosetType() string { return s.shosetType }

// GetIsValid returns isValid from Shoset.
func (s *Shoset) GetIsValid() bool { return s.isValid }

// GetIsPki returns isPki from Shoset.
func (s *Shoset) GetIsPki() bool { return s.isPki }

// GetListener returns listener from Shoset.
func (s *Shoset) GetListener() net.Listener { return s.listener }

// GetTlsConfigSingleWay returns tlsConfigSingleWay from Shoset.
func (s *Shoset) GetTlsConfigSingleWay() *tls.Config { return s.tlsConfigSingleWay }

// GetTlsConfigDoubleWay returns tlsConfigDoubleWay from Shoset.
func (s *Shoset) GetTlsConfigDoubleWay() *tls.Config { return s.tlsConfigDoubleWay }

// GetConnsByTypeArray returns an array of *ShosetConn known from a Shoset depending on a specific shosetType.
func (s *Shoset) GetConnsByTypeArray(shosetType string) []*ShosetConn {
	lNamesByType := s.LnamesByType.Keys(shosetType)
	var connsByType []*ShosetConn
	for _, lName := range lNamesByType {
		lNameMap, _ := s.ConnsByLname.smap.Load(lName)
		keys := Keys(lNameMap.(*sync.Map), ALL)
		for _, key := range keys {
			conn, _ := lNameMap.(*sync.Map).Load(key)
			connsByType = append(connsByType, conn.(*ShosetConn))
		}
	}
	return connsByType
}

// IsCertified returns true if path corresponds to an existing repertory which means that the Shoset has created its repertory and is certified .
func (s *Shoset) IsCertified(path string) bool {
	if fileExists(PATH_CA_PRIVATE_KEY) {
		s.SetIsPki(true)
	}
	return fileExists(path)
}

// SetBindAddress sets the bindAddress for a Shoset.
func (s *Shoset) SetBindAddress(bindAddress string) { s.bindAddress = bindAddress }

// SetIsValid sets the state for a Shoset.
func (s *Shoset) SetIsValid(state bool) { s.isValid = state }

// SetIsPki sets the state for a Shoset.
func (s *Shoset) SetIsPki(state bool) { s.isPki = state }

// SetListener sets the listener for a Shoset.
func (s *Shoset) SetListener(listener net.Listener) { s.listener = listener }

// deleteConn deletes a ShosetConn from ConnsByLname map from a Shoset
func (s *Shoset) deleteConn(connAddr, connLname string) {
	if connsByLname, _ := s.ConnsByLname.smap.Load(connLname); connsByLname != nil {
		if conn, _ := connsByLname.(*sync.Map).Load(connAddr); conn != nil {
			s.ConnsByLname.Delete(connLname, connAddr)
		}
	}
}

// NewShoset creates a new Shoset object.
// Initializes each fields, queues and handlers.
func NewShoset(logicalName, shosetType string) *Shoset {
	s := Shoset{
		Logger: log.With().Str("uuid", uuid.New()).Logger(),

		Context: make(map[string]interface{}),

		ConnsByLname:     new(MapSyncMap),
		LnamesByType:     new(MapSyncMap),
		LnamesByProtocol: new(MapSyncMap),
		ConnsSingleBool:  new(sync.Map),
		ConnsSingleConn:  new(sync.Map),

		RouteTable: new(sync.Map),

		logicalName: logicalName,
		shosetType:  shosetType,
		isValid:     true,
		isPki:       false,
		listener:    nil,

		tlsConfigSingleWay: &tls.Config{InsecureSkipVerify: true},
		tlsConfigDoubleWay: nil,

		Queue:    make(map[string]*msg.Queue),
		Handlers: make(map[string]MessageHandlers),
	}

	s.ConnsByLname.SetConfig(NewConfig())

	s.Queue["cfglink"] = msg.NewQueue()
	s.Handlers["cfglink"] = new(ConfigLinkHandler)

	s.Queue["cfgjoin"] = msg.NewQueue()
	s.Handlers["cfgjoin"] = new(ConfigJoinHandler)

	s.Queue["cfgbye"] = msg.NewQueue()
	s.Handlers["cfgbye"] = new(ConfigByeHandler)

	s.Queue["evt"] = msg.NewQueue()
	s.Handlers["evt"] = new(EventHandler)

	s.Queue["pkievt_TLSdoubleWay"] = msg.NewQueue()
	s.Handlers["pkievt_TLSdoubleWay"] = new(PkiEventHandler)

	s.Queue["pkievt_TLSsingleWay"] = msg.NewQueue()
	s.Handlers["pkievt_TLSsingleWay"] = new(PkiEventHandler)

	s.Queue["cmd"] = msg.NewQueue()
	s.Handlers["cmd"] = new(CommandHandler)

	//TODO MOVE TO GANDALF
	s.Queue["config"] = msg.NewQueue()
	s.Handlers["config"] = new(ConfigHandler)

	s.Logger.Debug().Str("lname", logicalName).Msg("shoset created")
	return &s
}

// String returns the formatted string of Shoset object in a pretty indented way.
func (s *Shoset) String() string {
	description := fmt.Sprintf("Shoset{\n\t- lName: %s,\n\t- bindAddr : %s,\n\t- type : %s, \n\t- isPki : %t, \n\t- ConnsByLname:", s.GetLogicalName(), s.GetBindAddress(), s.GetShosetType(), s.GetIsPki())

	connsByName := []*ShosetConn{}
	s.ConnsByLname.Iterate(
		func(key string, val interface{}) {
			connsByName = append(connsByName, val.(*ShosetConn))
		})
	sort.Slice(connsByName, func(i, j int) bool {
		return connsByName[i].GetRemoteAddress() < connsByName[j].GetRemoteAddress()
	})
	for _, connByName := range connsByName {
		description += fmt.Sprintf("\n\t\t* %s", connByName)
	}

	description += "\n\t- LnamesByProtocol:\n\t\t"
	keys := s.LnamesByProtocol.Keys(ALL)
	for _, key := range keys {
		description += key + "\t"
	}
	description += "\n\t"
	s.LnamesByProtocol.Iterate(
		func(key string, val interface{}) {
			description += fmt.Sprintf("\t%t", val.(bool))
		})

	description += "\n\t- LnamesByType:\n\t\t"
	keys = s.LnamesByType.Keys(ALL)
	for _, key := range keys {
		description += key + "\t\t"
	}
	description += "\n\t"
	s.LnamesByType.Iterate(
		func(key string, val interface{}) {
			description += fmt.Sprintf("\t%t", val.(bool))
		})
	//
	description += "\n\t- RouteTable (destination : {neighbour, distance, uuid}):\n\t\t"
	s.RouteTable.Range(
		func(key, val interface{}) bool {
			description += fmt.Sprintf("%v : %v", key, val)
			return true
		})

	description += "\n}\n"
	return description
}

// Bind assigns a local protocol address to the Shoset.
// Runs protocol on other Shosets if needed.
func (s *Shoset) Bind(address string) error {
	if err := s.ConnsByLname.GetConfig().ReadConfig(s.ConnsByLname.GetConfig().GetFileName()); err == nil {
		for _, remote := range s.ConnsByLname.cfg.GetSlice(PROTOCOL_JOIN) {
			s.Protocol(address, remote, PROTOCOL_JOIN)
		}
		for _, remote := range s.ConnsByLname.cfg.GetSlice(PROTOCOL_LINK) {
			s.Protocol(address, remote, PROTOCOL_LINK)
		}
	}

	ipAddress, err := GetIP(address)
	if err != nil {
		s.Logger.Error().Msg("couldn't set bindAddress : " + err.Error())
		return err
	}
	s.SetBindAddress(ipAddress)

	listener, err := net.Listen(CONNECTION_TYPE, DEFAULT_IP+strings.Split(ipAddress, ":")[1])
	if err != nil {
		s.Logger.Error().Msg("listen error : " + err.Error())
		return err
	}
	s.SetListener(listener)

	go s.handleBind()
	return nil
}

// handleBind listens on a specific port every occurring connection.
// Runs tls protocol to establish secured connection.
func (s *Shoset) handleBind() {
	defer s.GetListener().Close()
	for {
		acceptedConn, err := s.GetListener().Accept()
		if err != nil {
			s.Logger.Error().Msg("serverShoset accept error : " + err.Error())
			return
		}

		if exists, _ := s.ConnsSingleBool.Load(strings.Split(acceptedConn.RemoteAddr().String(), ":")[0]); exists != nil {
			tlsConnSingleWay := tls.Server(acceptedConn, s.tlsConfigSingleWay)
			singleWayConn, _ := NewShosetConn(s, acceptedConn.RemoteAddr().String(), IN)
			singleWayConn.UpdateConn(tlsConnSingleWay)
			go singleWayConn.RunInConnSingle(strings.Split(acceptedConn.RemoteAddr().String(), ":")[0])
		} else {
			tlsConnDoubleWay := tls.Server(acceptedConn, s.GetTlsConfigDoubleWay())
			doubleWayConn, _ := NewShosetConn(s, acceptedConn.RemoteAddr().String(), IN)
			doubleWayConn.UpdateConn(tlsConnDoubleWay)

			_, err = doubleWayConn.GetConn().Write([]byte(TLS_DOUBLE_WAY_TEST_WRITE + "\n"))
			if err == nil {
				go doubleWayConn.RunInConnDouble()
			} else {
				s.ConnsSingleBool.Store(strings.Split(acceptedConn.RemoteAddr().String(), ":")[0], true)
			}
		}
	}
}

// Protocol runs the suitable protocol handler.
// Inits certification if Shoset is not certified yet.
// Binds Shoset to the bindAddress if not bound yet.
func (s *Shoset) Protocol(bindAddress, remoteAddress, protocolType string) {
	s.Logger.Debug().Strs("params", []string{bindAddress, remoteAddress, protocolType}).Msg("protocol init")

	ipAddress, err := GetIP(bindAddress)
	if err != nil {
		s.Logger.Error().Msg("wrong IP format : " + err.Error())
		return
	}
	formattedIpAddress := strings.Replace(ipAddress, ":", "_", -1)
	formattedIpAddress = strings.Replace(formattedIpAddress, ".", "-", -1) // formats for filesystem to 127-0-0-1_8001 instead of 127.0.0.1:8001

	if !s.IsCertified(s.ConnsByLname.GetConfig().baseDir + formattedIpAddress) {
		s.Logger.Debug().Msg("ask certification")

		_, err = s.ConnsByLname.GetConfig().InitFolders(formattedIpAddress)
		if err != nil {
			s.Logger.Error().Msg("couldn't init folder: " + err.Error())
			return
		}

		s.ConnsByLname.GetConfig().SetFileName(formattedIpAddress)

		err = s.Certify(bindAddress, remoteAddress)
		if err != nil {
			return
		}
	}

	if s.GetBindAddress() == VOID {
		err := s.Bind(bindAddress)
		if err != nil {
			s.Logger.Error().Msg("couldn't set bindAddress : " + err.Error())
			return
		}
	}

	if remoteAddress == s.GetBindAddress() {
		s.Logger.Error().Msg("can't protocol on itself")
		return
	}

	protocolConn, _ := NewShosetConn(s, remoteAddress, OUT)
	cfg := msg.NewCfg(s.GetBindAddress(), s.GetLogicalName(), s.GetShosetType(), protocolType)
	go protocolConn.HandleConfig(cfg)
}
