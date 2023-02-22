package shoset

import (
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	uuid "github.com/kjk/betterguid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ditrit/shoset/msg"

	concurentData "github.com/ditrit/shoset/concurent_data"
	eventBus "github.com/ditrit/shoset/event_bus"
)

// MessageHandlers interface.
// Every handler must implement this interface
type MessageHandlers interface {
	Get(c *ShosetConn) (msg.Message, error)
	HandleDoubleWay(c *ShosetConn, message msg.Message) error
	Send(s *Shoset, m msg.Message)
	Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message
}

// Shoset : smart object based on network socket but with upgraded features
type Shoset struct {
	Logger zerolog.Logger // pretty logger

	Context map[string]interface{} // used for gandalf

	ConnsByLname     *MapSyncMap // map[lName]map[remoteAddress]*ShosetConn   connections by logical name
	LnamesByType     *MapSyncMap // map[shosetType]map[lName]bool used for gandalf
	LnamesByProtocol *MapSyncMap // map[protocolType]map[lName]bool logical names by protocol type
	ConnsSingleBool  *sync.Map   // map[ipAddress]bool ipAddresses waiting in singleWay to be handled for TLS double way
	ConnsSingleConn  *sync.Map   // map[ipAddress]*ShosetConn ShosetConns waiting in singleWay to be handled for TLS double way
	RouteTable       *sync.Map   // map[lName]*Route Route to another logical name

	RoutingEventBus eventBus.EventBus // When a route to a Lname is discovered, sends an event to everyone waiting for a route to this Lname
	// topic : discovered Lname

	bindAddress string       // address on which the Shoset is bound
	logicalName string       // logical name of the Shoset
	shosetType  string       // logical type of the shoset
	isValid     bool         // state of the Shoset - must be done differently in future review
	isPki       bool         // is the Shoset admin of network or not
	listener    net.Listener // generic network listener for stream-oriented protocols

	tlsConfigSingleWay *tls.Config // mutual authentication between client and server
	tlsConfigDoubleWay *tls.Config // client authenticate server

	Queue    map[string]*msg.Queue      // map for message queueing
	Handlers map[string]MessageHandlers // map for message handling

	MessageEventBus eventBus.EventBus // Sends an event to everyone waiting for a message of the type received
	// topic : MessageType

	Done chan bool // goroutines synchronization //Not used ?

	LaunchedProtocol concurentData.ConcurentSlice // List of IP addesses a Protocol was initiated with (but not yet finished)
	// The shoset is ready for use when the list is empty
	// Use WaitForProtocols() to wait for the shoset to be ready.

	mu sync.RWMutex
}

// GetBindAddress returns bindAddress from Shoset.
func (s *Shoset) GetBindAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bindAddress
}

// GetLogicalName returns lName from Shoset.
func (s *Shoset) GetLogicalName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logicalName
}

// GetShosetType returns shosetType from Shoset.
func (s *Shoset) GetShosetType() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.shosetType
}

// GetIsValid returns isValid from Shoset.
func (s *Shoset) GetIsValid() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isValid
}

// GetIsPki returns isPki from Shoset.
func (s *Shoset) GetIsPki() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isPki
}

// GetListener returns listener from Shoset.
func (s *Shoset) GetListener() net.Listener {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listener
}

// GetTlsConfigSingleWay returns tlsConfigSingleWay from Shoset.
func (s *Shoset) GetTlsConfigSingleWay() *tls.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tlsConfigSingleWay
}

// GetTlsConfigDoubleWay returns tlsConfigDoubleWay from Shoset.
func (s *Shoset) GetTlsConfigDoubleWay() *tls.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tlsConfigDoubleWay
}

// GetConnsByTypeArray returns an array of *ShosetConn known from a Shoset depending on a specific shosetType.
func (s *Shoset) GetConnsByTypeArray(shosetType string) []*ShosetConn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	lNamesByType := s.LnamesByType.Keys(shosetType)
	var connsByType []*ShosetConn
	for _, lName := range lNamesByType {
		lNameMap, _ := s.ConnsByLname.Load(lName)
		keys := Keys(lNameMap.(*sync.Map), ALL)
		for _, key := range keys {
			conn, _ := lNameMap.(*sync.Map).Load(key)
			connsByType = append(connsByType, conn.(*ShosetConn))
		}
	}
	return connsByType
}

// IsCertified returns true if path corresponds to an existing repertory which means that the Shoset has created its repertory and is certified.
func (s *Shoset) IsCertified(path string) bool {
	if fileExists(filepath.Join(path, PATH_CA_PRIVATE_KEY)) {
		s.SetIsPki(true)
	}
	return fileExists(path)
}

// SetBindAddress sets the bindAddress for a Shoset.
func (s *Shoset) SetBindAddress(bindAddress string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bindAddress = bindAddress
}

// SetIsValid sets the state for a Shoset.
func (s *Shoset) SetIsValid(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isValid = state
}

// SetIsPki sets the state for a Shoset.
func (s *Shoset) SetIsPki(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.isPki = state
}

// SetListener sets the listener for a Shoset.
func (s *Shoset) SetListener(listener net.Listener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listener = listener
}

// DeleteConn deletes a ShosetConn from ConnsByLname map from a Shoset
func (s *Shoset) DeleteConn(Lname, remoteAddress string) {
	// Lock shoset for the operation
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the ShosetCon exists
	conn, ok := s.ConnsByLname.LoadValueFromKeys(Lname, remoteAddress)
	if !ok {
		return
	}
	s.Logger.Debug().Msg("Deleting conn Lname : " + Lname + " IP : " + remoteAddress)

	c := conn.(*ShosetConn)

	s.ConnsByLname.DeleteValueFromKeys(Lname, remoteAddress)

	s.LnamesByProtocol.DeleteValueFromKeys(c.GetProtocol(), Lname)
	s.LnamesByType.DeleteValueFromKeys(c.GetRemoteShosetType(), Lname)
	s.ConnsByLname.DeleteValueFromKeys(Lname, remoteAddress)

	// Deletes from the config file
	s.ConnsByLname.cfg.DeleteFromKey(c.GetProtocol(), []string{c.GetRemoteAddress()})

	// Finds and deletes Routes using the deleted ShosetConn
	s.RouteTable.Range(
		func(key, val interface{}) bool {
			if val.(Route).GetNeighborConn() == c {
				c.GetShoset().RouteTable.Delete(key)
			}
			return true
		})
}

// Waits for every Conn initialised to be ready for use
func (s *Shoset) WaitForProtocols(timeout int) {
	s.Logger.Debug().Str("lname", s.GetLogicalName()).Msg("Waiting for all Protocol to complete on shoset.")
	err := s.LaunchedProtocol.WaitForEmpty(timeout)
	if err != nil {
		s.Logger.Error().Msg("Failed to establish connection to some adresses (timeout) : " + s.LaunchedProtocol.String())
	} else {
		s.Logger.Debug().Str("lname", s.GetLogicalName()).Msg("All Protocols done on shoset.")
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
		RouteTable:       new(sync.Map),

		RoutingEventBus: eventBus.NewEventBus(),

		logicalName: logicalName,
		shosetType:  shosetType,
		isValid:     true,
		isPki:       false,
		listener:    nil,

		tlsConfigSingleWay: &tls.Config{InsecureSkipVerify: true},
		tlsConfigDoubleWay: nil,

		Queue:           make(map[string]*msg.Queue),
		Handlers:        make(map[string]MessageHandlers),
		MessageEventBus: eventBus.NewEventBus(),

		LaunchedProtocol: concurentData.NewConcurentSlice(),
	}

	s.ConnsByLname.SetConfig(NewConfig(s.logicalName))

	//s.Queue["cfglink"] = msg.NewQueue() // Not neeeded
	s.Handlers["cfglink"] = new(ConfigLinkHandler)

	//s.Queue["cfgjoin"] = msg.NewQueue() // Not neeeded
	s.Handlers["cfgjoin"] = new(ConfigJoinHandler)

	//s.Queue["cfgbye"] = msg.NewQueue() // Not neeeded
	s.Handlers["cfgbye"] = new(ConfigByeHandler)

	s.Queue["evt"] = msg.NewQueue()
	s.Handlers["evt"] = new(EventHandler)

	s.Queue["pkievt_TLSdoubleWay"] = msg.NewQueue()
	s.Handlers["pkievt_TLSdoubleWay"] = new(PkiEventHandler)

	s.Queue["pkievt_TLSsingleWay"] = msg.NewQueue()
	s.Handlers["pkievt_TLSsingleWay"] = new(PkiEventHandler)

	// s.Queue["routingEvent"] = msg.NewQueue() // Not neeeded
	s.Handlers["routingEvent"] = new(RoutingEventHandler)

	s.Queue["forwardAck"] = msg.NewQueue()
	s.Handlers["forwardAck"] = new(ForwardAcknownledgeHandler)

	s.Queue["simpleMessage"] = msg.NewQueue()
	s.Handlers["simpleMessage"] = new(SimpleMessageHandler)

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
	s.mu.RLock()
	defer s.mu.RUnlock()
	description := fmt.Sprintf("Shoset{\n\t- lName: %s,\n\t- bindAddr : %s,\n\t- type : %s, \n\t- isPki : %t", s.GetLogicalName(), s.GetBindAddress(), s.GetShosetType(), s.GetIsPki())

	description += ", \n\t- LaunchedProtocol : " + s.LaunchedProtocol.String()

	description += ", \n\t- ConnsByLname : "
	connsByName := []*ShosetConn{}
	s.ConnsByLname.Iterate(
		func(lname string, ipAddress string, val interface{}) {
			connsByName = append(connsByName, val.(*ShosetConn))
		})
	sort.Slice(connsByName, func(i, j int) bool {
		return connsByName[i].GetRemoteAddress() < connsByName[j].GetRemoteAddress()
	})
	for _, connByName := range connsByName {
		description += fmt.Sprintf("\n\t\t&%s", connByName)
	}

	description += "\n\t- LnamesByProtocol:\n\t"
	description += s.LnamesByProtocol.String()
	description += "\n\t- LnamesByType:\n\t"
	description += s.LnamesByType.String()

	description += "\n\t- RouteTable (destination : {neighbor, Conn to Neighbor, distance, uuid, timestamp}):\n\t\t"
	routeTable := map[string]Route{}
	routeTableKey := []string{}
	s.RouteTable.Range(
		func(key, val interface{}) bool {
			routeTable[key.(string)] = val.(Route)
			routeTableKey = append(routeTableKey, key.(string))
			return true
		})
	sort.Strings(routeTableKey)
	for _, routeKey := range routeTableKey {
		description += fmt.Sprintf("\n\t\t%v : %v", routeKey, routeTable[routeKey])
	}

	description += "\n}\n"
	return description
}

// Bind assigns a local protocol address to the Shoset.
// Runs protocol on other Shosets if needed.
func (s *Shoset) Bind(address string) error {
	ipAddress, err := GetIP(address)
	if err != nil {
		s.Logger.Error().Msg("couldn't set bindAddress : " + err.Error())
		return err
	}
	s.SetBindAddress(ipAddress)

	err = s.ConnsByLname.GetConfig().ReadConfig(s.ConnsByLname.GetConfig().GetFileName())
	if err == nil {
		for _, remote := range s.ConnsByLname.cfg.GetSlice(PROTOCOL_JOIN) {
			s.Protocol(address, remote, PROTOCOL_JOIN)
		}
		for _, remote := range s.ConnsByLname.cfg.GetSlice(PROTOCOL_LINK) {
			s.Protocol(address, remote, PROTOCOL_LINK)
		}
	}

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

			_, err = doubleWayConn.GetConn().Write([]byte(TLS_DOUBLE_WAY_TEST_WRITE + "\n")) // Crashes when launching 2 shosets at the same time
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

	s.mu.Lock()
	s.ConnsByLname.GetConfig().SetFileName(formattedIpAddress)
	s.mu.Unlock()

	if !s.IsCertified(filepath.Join(s.ConnsByLname.GetConfig().baseDirectory, formattedIpAddress)) {
		s.Logger.Debug().Msg("ask certification")

		_, err = s.ConnsByLname.GetConfig().InitFolders(formattedIpAddress)
		if err != nil {
			s.Logger.Error().Msg("couldn't init folder: " + err.Error())
			return
		}
		err = s.Certify(bindAddress, remoteAddress)
		if err != nil {
			return
		}
	} else {
		// Loads certificats from the folder
		err = s.SetUpDoubleWay()
		if err != nil {
			s.Logger.Error().Msg(err.Error())
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

	if IP, _ := GetIP(remoteAddress); IP == s.GetBindAddress() {
		s.Logger.Error().Msg("can't protocol on itself : " + IP)
		return
	}

	if remoteAddress != VOID {
		protocolConn, _ := NewShosetConn(s, remoteAddress, OUT)
		cfg := msg.NewConfigProtocol(s.GetBindAddress(), s.GetLogicalName(), s.GetShosetType(), protocolType)

		s.LaunchedProtocol.AppendToConcurentSlice(protocolConn.GetRemoteAddress()) // Adds remote adress to the list of initiated but not ready connexion adresses
		go protocolConn.HandleConfig(cfg)
	}
}

func (s *Shoset) EndProtocol(Lname, remoteAddress string) {
	// Finds the ShosetConn in the list
	var c *ShosetConn
	if conn, ok := s.ConnsByLname.LoadValueFromKeys(Lname, remoteAddress); !ok {
		s.Logger.Error().Msg("No Existing connection to Lname : " + Lname + " IP : " + remoteAddress + ", no connection to end.")
		return
	} else {
		c = conn.(*ShosetConn)
	}

	cfg := msg.NewConfigProtocol(s.GetBindAddress(), s.GetLogicalName(), s.GetShosetType(), DELETE)

	err := c.GetWriter().SendMessage(*cfg)
	if err != nil {
		c.Logger.Error().Msg("couldn't send cfg: " + err.Error())
		return
		//Retry ? Acknowledge ?
	}

	s.LnamesByProtocol.AppendToKeys(PROTOCOL_EXIT, Lname, true) //Bye ou delete ?
	s.DeleteConn(Lname, remoteAddress)
}

// ######## Route and forwarding : ########

// Forward messages destined to another Lname to the next step on the Route
func (s *Shoset) forwardMessage(m msg.Message) {
	masterTimeout := time.NewTimer(time.Duration(MASTER_SEND_TIMEOUT) * time.Second)
	tryNumber := 0

	for {
		if route, ok := s.RouteTable.Load(m.GetDestinationLname()); ok { // There is a known Route to the destination Lname
			s.Logger.Debug().Msg(s.GetLogicalName() + " is sending a message to " + m.GetDestinationLname() + " through " + route.(Route).GetNeighbour() + " IP : " + route.(Route).GetNeighborConn().GetRemoteAddress() + " .")

			// Forward message
			err := route.(Route).GetNeighborConn().GetWriter().SendMessage(m)
			if err != nil {
				s.Logger.Error().Msg("Couldn't send forwarded message : " + err.Error())
			}

			// Wait for Acknowledge
			forwardAck := s.Wait("forwardAck", map[string]string{"UUID": m.GetUUID()}, TIMEOUT_ACK, nil)
			if forwardAck == nil {
				s.Logger.Warn().Msg("Forward message : Failed to forward message destined to " + m.GetDestinationLname() + " Forward Acknowledge not received. (retrying)")

				// Invalidate route
				s.RouteTable.Delete(m.GetDestinationLname())

				// Reroute network
				routing := msg.NewRoutingEvent(s.GetLogicalName(), true, 0, "")
				s.Send(routing)

				tryNumber++
				if tryNumber > MAX_FORWARD_TRY {
					s.Logger.Warn().Msg("Forward message : Failed to forward message destined to " + m.GetDestinationLname() + ". Max number of attemp exceeded. (Giving up)")
					return
				} else {
					continue
				}
			}
			return

		} else { // There is no known Route to the destination Lname -> Wait for one to be available
			s.Logger.Warn().Msg("Forward message : Failed to forward message destined to " + m.GetDestinationLname() + ". (no route) (waiting for correct route)")

			// Reroute network
			routing := msg.NewRoutingEvent(s.GetLogicalName(), true, 0, "")
			s.Send(routing)

			// Creation channel
			chRoute := make(chan interface{})
			s.RoutingEventBus.Subscribe(m.GetDestinationLname(), chRoute)

			// Check if the route did not become available while you were preparing the channel
			_, ok := s.RouteTable.Load(m.GetDestinationLname())
			if ok {
				continue
			}
			select {
			case <-chRoute:
				s.RoutingEventBus.UnSubscribe(m.GetDestinationLname(), chRoute)
				break
			case <-masterTimeout.C:
				s.Logger.Error().Msg("Couldn't send forwarded message : " + "Timed out before correct route discovery. (Waited to long for the route)")
				return
			}
		}
	}
}

func (s *Shoset) SaveRoute(c *ShosetConn, routingEvt *msg.RoutingEvent) {
	s.Logger.Debug().Msg(s.GetLogicalName() + " is saving a route to " + routingEvt.GetOrigin() + " through " + c.GetRemoteLogicalName() + " IP : " + c.GetRemoteAddress())

	s.RouteTable.Store(routingEvt.GetOrigin(), NewRoute(c.GetRemoteLogicalName(), c, routingEvt.GetNbSteps(), routingEvt.GetUUID(), routingEvt.GetRerouteTimestamp()))

	// Send NewRouteEvent
	s.RoutingEventBus.Publish(routingEvt.GetOrigin(), true)

	// Reroutes self to try to take advantage of the change in the network that triggered the Save but with the same timestamp and ID to avoid triggering it over and over.
	reRouting := msg.NewRoutingEvent(c.GetLocalLogicalName(), false, routingEvt.GetRerouteTimestamp(), routingEvt.GetUUID())
	s.Send(reRouting)

	// Rebroadcast Routing event
	routingEvt.SetNbSteps(routingEvt.GetNbSteps() + 1)
	s.Send(routingEvt)
}

// ######## Send and Receice Messages : ########

// Finds the correct send function for the type of message using the handler and call it.
func (s *Shoset) Send(msg msg.Message) {
	if msgType := msg.GetMessageType(); contains(SENDABLE_TYPES, msgType) {
		s.Handlers[msgType].Send(s, msg)
	} else {
		s.Logger.Error().Msg("Trying to send an invalid message type or message of a type without a Send function. Message type : " + msgType)
	}
}

//Wait for message
//args for event("evt") type : map[string]string{"topic": "topic_name", "event": "event_name"}
//Leave iterator at nil if you don't want to supply it yourself. (avoids reading multiple time the same message)
func (s *Shoset) Wait(msgType string, args map[string]string, timeout int, iterator *msg.Iterator) msg.Message {
	if !contains(RECEIVABLE_TYPES, msgType) {
		s.Logger.Error().Msg("Trying to receive an invalid message type or message of a type without a Wait function. Message type : " + msgType)
		return nil
	}

	if iterator == nil {
		iterator = msg.NewIterator(s.Queue[msgType])
	}

	event := s.Handlers[msgType].Wait(s, iterator, args, timeout)

	if event == nil {
		return nil
	} else {
		return *(event)
	}

}
