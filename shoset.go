package shoset

import (
	"crypto/tls"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	concurentData "github.com/ditrit/shoset/concurent_data"
	eventBus "github.com/ditrit/shoset/event_bus"
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
	RouteTable       *sync.Map   // map[lName]*Route Route to another logical name

	//NewRouteEvent   chan string // Notify of the discovery of a new route
	RoutingEventBus eventBus.EventBus // When a route to a Lname is discovered, sends an event to everyone waiting for a route to this Lname
	// topic : discovveredLname

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

	Done chan bool // goroutines synchronization

	LaunchedProtocol concurentData.ConcurentSlice // List of IP addesses a Protocol was initiated with (but not yet finished)
	// The shoset is ready to use when the list is empty
	// Use WaitForProtocols() to wait for the shoset to be ready.
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
		lNameMap, _ := s.ConnsByLname.Load(lName)
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
	if connsByLname, _ := s.ConnsByLname.Load(connLname); connsByLname != nil {
		if conn, _ := connsByLname.(*sync.Map).Load(connAddr); conn != nil {
			//fmt.Println("ConnsByLname : ", s.ConnsByLname)
			s.ConnsByLname.DeleteConfig(connLname, connAddr)
		}
	}
}

// Wait for every Conn initialised to be ready for use
// Add timeout
func (s *Shoset) WaitForProtocols(timeout int) {
	fmt.Println("Waiting for all Protocol to complete on shoset", s.GetLogicalName())
	//s.waitGroupProtocol.Wait()
	//fmt.Println("s.LaunchedProtocol : ", s.LaunchedProtocol.String())
	err := s.LaunchedProtocol.WaitForEmpty(timeout)
	if err != nil {
		s.Logger.Error().Msg("Failed to establish connection to some adresses (timeout) : " + s.LaunchedProtocol.String())
		//fmt.Println("Shoset after Protocol : ",s)
	} else {
		fmt.Println("All Protocols done for ", s.GetLogicalName())
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

		//NewRouteEvent: make(chan string),
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

	s.ConnsByLname.SetConfig(NewConfig(s.logicalName)) // Add baseDir parameter

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

	// s.Queue["routingEvent"] = msg.NewQueue() // Not neeeded,
	s.Handlers["routingEvent"] = new(RoutingEventHandler)

	s.Queue["forwardAck"] = msg.NewQueue()
	s.Handlers["forwardAck"] = new(forwardAcknownledgeHandler)

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

	description += "\n\t- LnamesByProtocol:\n\t"
	description += s.LnamesByProtocol.String()
	description += "\n\t- LnamesByType:\n\t"
	description += s.LnamesByType.String()

	description += "\n\t- RouteTable (destination : {neighbor, Conn to Neighbor, distance, uuid, timestamp}):\n\t\t"
	s.RouteTable.Range(
		func(key, val interface{}) bool {
			description += fmt.Sprintf("%v : %v \n\t\t", key, val)
			return true
		})

	description += "\n}\n"
	return description
}

// Bind assigns a local protocol address to the Shoset.
// Runs protocol on other Shosets if needed.
func (s *Shoset) Bind(address string) error {
	fmt.Println("Bind")

	ipAddress, err := GetIP(address)
	if err != nil {
		s.Logger.Error().Msg("couldn't set bindAddress : " + err.Error())
		return err
	}
	s.SetBindAddress(ipAddress)
	//fmt.Println("(Bind) BindAdress : ", s.GetBindAddress())

	//fmt.Println("FileName : ", s.ConnsByLname.GetConfig().GetFileName())

	err = s.ConnsByLname.GetConfig().ReadConfig(s.ConnsByLname.GetConfig().GetFileName())

	if err == nil {
		fmt.Println("Sclice JOIN : ", s.ConnsByLname.cfg.GetSlice(PROTOCOL_JOIN))
		for _, remote := range s.ConnsByLname.cfg.GetSlice(PROTOCOL_JOIN) {
			s.Protocol(address, remote, PROTOCOL_JOIN)
		}
		fmt.Println("Sclice LINK : ", s.ConnsByLname.cfg.GetSlice(PROTOCOL_LINK))
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

			//fmt.Println("(handleBind) doubleWayConn.GetConn() : ",doubleWayConn.GetConn())

			_, err = doubleWayConn.GetConn().Write([]byte(TLS_DOUBLE_WAY_TEST_WRITE + "\n")) // Crash
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
	fmt.Println("PROTOCOL !!!!")
	s.Logger.Debug().Strs("params", []string{bindAddress, remoteAddress, protocolType}).Msg("protocol init")

	ipAddress, err := GetIP(bindAddress)
	if err != nil {
		s.Logger.Error().Msg("wrong IP format : " + err.Error())
		return
	}
	formattedIpAddress := strings.Replace(ipAddress, ":", "_", -1)
	formattedIpAddress = strings.Replace(formattedIpAddress, ".", "-", -1) // formats for filesystem to 127-0-0-1_8001 instead of 127.0.0.1:8001

	//fmt.Println("File Path : ", s.ConnsByLname.GetConfig().baseDir+formattedIpAddress, "Certified : ", s.IsCertified(s.ConnsByLname.GetConfig().baseDir+formattedIpAddress))

	s.ConnsByLname.GetConfig().SetFileName(formattedIpAddress)

	if !s.IsCertified(s.ConnsByLname.GetConfig().baseDir + formattedIpAddress) {
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
		err = s.SetUpDoubleWay()
		if err != nil {
			s.Logger.Error().Msg(err.Error())
			return
		}
	}

	//fmt.Println("(Protocol) BindAdress : ", s.GetBindAddress())
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

	if remoteAddress != "" {
		fmt.Println("PROTOCOL ON  REMOTE !!!!")
		protocolConn, _ := NewShosetConn(s, remoteAddress, OUT)
		cfg := msg.NewConfigProtocol(s.GetBindAddress(), s.GetLogicalName(), s.GetShosetType(), protocolType)

		//s.waitGroupProtocol.Add(1)
		s.LaunchedProtocol.AppendToConcurentSlice(protocolConn.GetRemoteAddress()) // Adds remote adress to the list of initiated but not ready connexion adresses
		go protocolConn.HandleConfig(cfg)

		//fmt.Println("Certificates singleWay : ", s.tlsConfigSingleWay)
		//fmt.Println("Certificates doubleWay : ", s.tlsConfigDoubleWay)
	}
}

// Forward messages destined to another Lname to the next step on the Route
func (s *Shoset) forwardMessage(m msg.Message) {
	masterTimeout := time.NewTimer(time.Duration(MASTER_SEND_TIMEOUT) * time.Second)

	tryNumber := 0

	for {
		route, ok := s.RouteTable.Load(m.GetDestinationLname())
		if ok { // There is a known Route to the destination Lname
			//fmt.Println("((SimpleMessageHandler) Send) ", s.GetLogicalName(), " is sending a message to ", m.GetDestinationLname(), "through ", route.(Route).neighbour, ".")

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
				routing := msg.NewRoutingEvent(s.GetLogicalName(), "")
				s.Send(routing)

				tryNumber++
				if tryNumber > MAX_FORWARD_TRY {
					return
				} else {
					continue
				}
			}
			fmt.Println("(ForwardAck) Message received : ", forwardAck)

			return

		} else { // There is no known Route to the destination Lname -> Wait for one to be available
			s.Logger.Warn().Msg("Forward message : Failed to forward message destined to " + m.GetDestinationLname() + ". (no route) (waiting for correct route")

			// Reroute network
			routing := msg.NewRoutingEvent(s.GetLogicalName(), "")
			s.Send(routing)

			//retry: // Puts a label on the loop to break out of it from inside the select
			// Wait for a route to the destination to be dicovered
			// Timeout :
			// - Since the last addition to the Routetable
			// - Since the bigenning of the wait

			//for {

			// Creation channel
			chRoute := make(chan interface{})

			// defer un Unsub

			// Inscription channel
			s.RoutingEventBus.Subscribe(m.GetDestinationLname(), chRoute)

			defer s.RoutingEventBus.UnSubscribe(m.GetDestinationLname(), chRoute) // ?

			fmt.Println(s.GetLogicalName(), " is wating for a route to ", m.GetDestinationLname())

			// Possibilité de trouver la bonne route avant d'être abonné ? (Refaire le load avant d'attendre ?)
			_, ok := s.RouteTable.Load(m.GetDestinationLname())
			if ok {
				continue
			}
			select {
			case <-chRoute:
				//fmt.Println(s.GetLogicalName(), "Received NewRouteEvent for ", m.GetDestinationLname())
				break
				// if Lname == m.GetDestinationLname() {
				// 	break retry
				// }
			// case <-time.After(time.Duration(NO_MESSAGE_ROUTE_TIMEOUT) * time.Second):
			// 	// When the message is sent when this is not receiving, it is not sent and never received
			// 	// Retry anyway after some time, maybe the Event was missed

			// 	s.Logger.Debug().Msg("Timed out before correct route discovery. (no recent NewRouteEvent) (retrying)")
			// 	return

			case <-masterTimeout.C:
				s.Logger.Error().Msg("Couldn't send forwarded message : " + "Timed out before correct route discovery. (Waited to long for the route)")
				return
			}
			//}
			fmt.Println("Retrying forward")
		}
	}
}

func (s *Shoset) SaveRoute(c *ShosetConn, routingEvt *msg.RoutingEvent) {
	s.RouteTable.Store(routingEvt.GetOrigin(), NewRoute(c.GetRemoteLogicalName(), c, routingEvt.GetNbSteps(), routingEvt.GetUUID(), routingEvt.Timestamp))

	// Send NewRouteEvent
	s.RoutingEventBus.Publish(routingEvt.GetOrigin(), true) // Sent data is not used

	// select {
	// case s.NewRouteEvent <- routingEvt.GetOrigin():
	// 	fmt.Println(c.GetLocalLogicalName(), " is Sending NewRouteEvent for ", routingEvt.GetOrigin())
	// default:
	// 	//fmt.Println("Nobody is waiting for NewRouteEvent")
	// }

	reRouting := msg.NewRoutingEvent(c.GetLocalLogicalName(), routingEvt.GetUUID())
	s.Send(reRouting)

	// Rebroadcast Routing event
	routingEvt.SetNbSteps(routingEvt.GetNbSteps() + 1)
	s.Send(routingEvt)
}

// ######## Send and Receice Messages : ########

// Find the correct send function for the type of message using the handler and call it
func (s *Shoset) Send(msg msg.Message) { //Use pointer for msg ?
	s.Handlers[msg.GetMessageType()].Send(s, msg)
}

//Wait for message
//args for event("evt") type : map[string]string{"topic": "topic_name", "event": "event_name"}
//Leave iterator at nil if you don't want to supply it yourself (avoid reading multiple time the same message)
func (s *Shoset) Wait(msgType string, args map[string]string, timeout int, iterator *msg.Iterator) msg.Message {
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
