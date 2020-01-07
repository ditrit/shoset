package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	//	uuid "github.com/kjk/betterguid"

	"./msg"
)

//Chaussette : client gandalf Socket
type Chaussette struct {
	//	id          string
	connsByAddr map[string]*ChaussetteConn
	connsByName map[string]map[string]*ChaussetteConn
	brothers    map[string]bool
	logicalName string
	done        chan bool
	bindAddress string
	m           sync.RWMutex
	qEvents     *msg.Queue
	qCommands   *msg.Queue
	qReplies    *msg.Queue
	qConfigs    *msg.Queue
	tlsConfig   *tls.Config
	tlsServerOK bool
	brothersIn  map[string]bool
}

var certPath = "./cert.pem"
var keyPath = "./key.pem"

// NewChaussette : constructor
func NewChaussette(logicalName string) *Chaussette {
	c := new(Chaussette)
	//	c.id = uuid.New()
	c.logicalName = logicalName
	c.connsByAddr = make(map[string]*ChaussetteConn)
	c.connsByName = make(map[string]map[string]*ChaussetteConn)
	c.brothers = make(map[string]bool)
	c.qEvents = msg.NewQueue()
	c.qCommands = msg.NewQueue()
	c.qReplies = msg.NewQueue()
	c.qConfigs = msg.NewQueue()
	c.brothersIn = make(map[string]bool)

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

// ChaussetteConn : client connection
type ChaussetteConn struct {
	socket            *tls.Conn
	remoteLogicalName string
	remoteBindAddress string
	direction         string
	address           string
	chaussette        *Chaussette
	rb                *msg.Reader
	wb                *msg.Writer
}

//NewHandshakeMessage : Build a config Message
func (c *Chaussette) NewHandshakeMessage() *msg.Config {
	lName := c.logicalName
	listes := map[string]map[string][]string{"in": map[string][]string{}, "out": map[string][]string{}}
	for _, conn := range c.connsByAddr {
		dir := conn.direction
		if listes[dir] == nil {
			return nil
		}
		rlName := conn.remoteLogicalName
		addr := conn.address
		if rlName != "" {
			if listes[dir][rlName] == nil {
				listes[dir][rlName] = []string{addr}
			} else {
				listes[dir][rlName] = append(listes[dir][rlName], addr)
			}
		}
	}
	return msg.NewHandshake(c.bindAddress, lName, listes)
}

//NewInstanceMessage : Build a config Message
func (c *Chaussette) NewInstanceMessage(address string, logicalName string) *msg.Config {
	return msg.NewInstance(address, logicalName)
}

//NewConnectToMessage : Build a config Message
func (c *Chaussette) NewConnectToMessage(address string) *msg.Config {
	return msg.NewConnectTo(address)
}

func getIP(address string) (string, error) {
	parts := strings.Split(address, ":")
	if len(parts) != 2 {
		return "", errors.New("address '" + address + "should respect the format hots_name_or_ip:port")
	}
	hostIps, err := net.LookupHost(parts[0])
	if err != nil || len(hostIps) == 0 {
		return "", errors.New("address '" + address + "' can not be resolved")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", errors.New("'" + parts[1] + "' is not a port number")
	}
	if port < 1 || port > 65535 {
		return "", errors.New("'" + parts[1] + "' is not a valid port number")
	}
	ipaddr := hostIps[0] + ":" + parts[1]
	return ipaddr, nil
}

//Connect : Connect to another Chaussette
func (c *Chaussette) Connect(address string) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.chaussette = c
	conn.direction = "out"
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := getIP(address)
	if err != nil {
		return nil, err
	}
	conn.address = ipAddress
	go conn.runOutConn(conn.address)
	return conn, nil
}

// RunOutConn : handler for the socket
func (c *ChaussetteConn) runOutConn(address string) {
	myConfig := c.chaussette.NewHandshakeMessage()
	for {
		c.chaussette.setConn(address, c)
		conn, err := tls.Dial("tcp", c.address, c.chaussette.tlsConfig)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)

			// receive messages
			for {
				if c.remoteLogicalName == "" {
					c.SendConfig(myConfig)
				}
				c.receiveMsg()
			}
		}
	}
}

func (c *Chaussette) deleteConn(connAddr string) {
	c.m.Lock()
	conn := c.connsByAddr[connAddr]
	if conn != nil {
		logicalName := conn.remoteLogicalName
		if c.connsByName[logicalName] != nil {
			delete(c.connsByName[logicalName], connAddr)
		}
	}
	delete(c.connsByAddr, connAddr)
	c.m.Unlock()
}

func (c *Chaussette) setConn(connAddr string, conn *ChaussetteConn) {
	if conn != nil {
		c.m.Lock()
		c.connsByAddr[connAddr] = conn
		logicalName := conn.remoteLogicalName
		if logicalName != "" {
			if c.connsByName[logicalName] == nil {
				c.connsByName[logicalName] = make(map[string]*ChaussetteConn)
			}
			c.connsByName[logicalName][connAddr] = conn
		}
		c.m.Unlock()
	}
}

func (c *ChaussetteConn) setRemoteLogicalName(logicalName string) {
	if logicalName != "" {
		c.remoteLogicalName = logicalName
		if c.chaussette.connsByName[logicalName] == nil {
			c.chaussette.connsByName[logicalName] = make(map[string]*ChaussetteConn)
		}
		c.chaussette.connsByName[logicalName][c.address] = c
	}
}

func (c *ChaussetteConn) setRemoteBindAddress(bindAddress string) {
	if bindAddress != "" {
		c.remoteBindAddress = bindAddress
	}
}

//Bind : Connect to another Chaussette
func (c *Chaussette) Bind(address string) error {
	if c.bindAddress != "" {
		fmt.Println("Chaussette already bound")
		return errors.New("Chaussette already bound")
	}
	if c.tlsServerOK == false {
		fmt.Println("TLS configuration not OK (certificate not found / loaded)")
		return errors.New("TLS configuration not OK (certificate not found / loaded)")
	}
	ipAddress, err := getIP(address)
	if err != nil {
		return err
	}
	c.bindAddress = ipAddress
	fmt.Printf("Bind : handleBind adress %s", ipAddress)
	go c.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (c *Chaussette) handleBind() error {
	listener, err := net.Listen("tcp", c.bindAddress)
	if err != nil {
		fmt.Println("Failed to bind:", err.Error())
		return err
	}
	defer listener.Close()

	for {
		unencConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("serverChaussette accept error: %s", err)
			break
		}
		tlsConn := tls.Server(unencConn, c.tlsConfig)
		conn, _ := c.inboudConn(tlsConn)
		fmt.Printf("Chaussette : accepted from %s", conn.address)
		go conn.runInConn()
	}
	return nil
}

//inboudConn : Add a new connection from a client
func (c *Chaussette) inboudConn(tlsConn *tls.Conn) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.socket = tlsConn
	conn.direction = "in"
	conn.chaussette = c
	conn.address = tlsConn.RemoteAddr().String()
	c.setConn(conn.address, conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	return conn, nil
}

// runInbound : handler for the connection
func (c *ChaussetteConn) runInConn() {
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)
	myConfig := c.chaussette.NewHandshakeMessage()

	// receive messages
	for {
		if c.remoteLogicalName == "" {
			c.SendConfig(myConfig)
		}
		err := c.receiveMsg()
		if err != nil {
			return
		}
	}

}

// handleConfigMessages :
func (c *ChaussetteConn) handleConfigMessages() error {
	var cfg msg.Config
	err := c.rb.ReadConfig(&cfg)
	switch cfg.GetCommandName() {
	case "handshake":
		if c.remoteLogicalName == "" {
			c.setRemoteLogicalName(cfg.GetLogicalName())
		}
		if c.remoteBindAddress == "" {
			c.setRemoteBindAddress(cfg.GetBindAddress())
		}
		if c.direction == "in" {
			if c.remoteBindAddress != "" {
				cfgNewInstance := c.chaussette.NewInstanceMessage(c.remoteBindAddress, c.remoteLogicalName)
				fmt.Printf("\n newInstance : %#v\n", cfgNewInstance)
				for _, conn := range c.chaussette.connsByName[c.remoteLogicalName] {
					if conn.direction == "in" {
						conn.SendConfig(cfgNewInstance)
					}
				}
				myBrothers := cfg.Conns["out"][c.chaussette.logicalName]
				for _, myBrother := range myBrothers {
					c.chaussette.brothers[myBrother] = true
				}
				for addrBrother := range c.chaussette.brothers {
					cfgConnectTo := c.chaussette.NewConnectToMessage(addrBrother)
					c.SendConfig(cfgConnectTo)
				}
			}
		}
		if c.direction == "out" {
			/*
				connBrothers := cfg.Conns["in"][c.chaussette.logicalName]
				for _, connBrother := range connBrothers {
					if c.chaussette.connsByAddr[connBrother] == nil {
						c.chaussette.Connect(connBrother)
					}
				}*/

		}
	case "newInstance":
		if c.direction == "out" {
			cfgConnectTo := c.chaussette.NewConnectToMessage(cfg.Address)
			for _, conn := range c.chaussette.connsByAddr {
				if conn.direction == "in" {
					conn.SendConfig(cfgConnectTo)
				}
			}
		}
	case "connectTo":
		if c.direction == "out" {

			if c.chaussette.connsByAddr[cfg.Address] == nil {
				fmt.Printf("\n %s.connectTo : %s\n", c.chaussette.bindAddress, cfg.Address)
				c.chaussette.Connect(cfg.Address)
			}
		}
	}
	return err
}

func (c *ChaussetteConn) receiveMsg() error {
	// read message type
	msgType, err := c.rb.ReadString()
	switch {
	case err == io.EOF:
		c.chaussette.deleteConn(c.address)
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		c.chaussette.deleteConn(c.address)
		return errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")

	// read message data
	//fmt.Printf("Read message and push if into buffer")
	switch msgType {
	case "evt":
		var evt msg.Event
		err = c.rb.ReadEvent(&evt)
		c.chaussette.qEvents.Push(evt)
	case "cmd":
		var cmd msg.Command
		err = c.rb.ReadCommand(&cmd)
		c.chaussette.qCommands.Push(cmd)
	case "rep":
		var rep msg.Reply
		err = c.rb.ReadReply(&rep)
		c.chaussette.qReplies.Push(rep)
	case "cfg":
		c.handleConfigMessages()
	default:
		c.chaussette.deleteConn(c.address)
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	if err != nil {
		c.chaussette.deleteConn(c.address)
		return errors.New("receiveMsg : unable to decode a message of type  " + msgType)
	}
	return err
}

// SendEvent : send an event
func (c *ChaussetteConn) SendEvent(evt *msg.Event) {
	c.wb.WriteString("evt")
	c.wb.WriteEvent(*evt)
}

// SendEvent : send an event...
// event is sent on each connection
func (c *Chaussette) SendEvent(evt *msg.Event) {
	fmt.Print("Sending event.\n")
	for _, conn := range c.connsByAddr {
		conn.SendEvent(evt)
	}
}

// SendCommand : Send a message
func (c *ChaussetteConn) SendCommand(cmd *msg.Command) {
	c.wb.WriteString("cmd")
	c.wb.WriteCommand(*cmd)
}

// SendCommand : Send a message
// todo : manage routing
//    identify relevant targets (routing info matches identity)
//    then try on each instance until success
func (c *Chaussette) SendCommand(cmd *msg.Command) {
	fmt.Print("Sending command.\n")
	for _, conn := range c.connsByAddr {
		conn.SendCommand(cmd)
	}
}

// SendReply :
func (c *ChaussetteConn) SendReply(rep *msg.Reply) {
	c.wb.WriteString("rep")
	c.wb.WriteReply(*rep)
}

// SendReply :
func (c *Chaussette) SendReply(rep *msg.Reply) {
	fmt.Print("Sending reply.\n")
	for _, conn := range c.connsByAddr {
		conn.SendReply(rep)
	}
}

// SendConfig :
func (c *ChaussetteConn) SendConfig(cfg *msg.Config) {
	fmt.Print("Sending config.\n")
	c.wb.WriteString("cfg")
	c.wb.WriteConfig(*cfg)
}

// SendConfig :
func (c *Chaussette) SendConfig(cfg *msg.Config) {
	fmt.Print("Sending configuration.\n")
	for _, conn := range c.connsByAddr {
		conn.SendConfig(cfg)
	}
}

// WaitEvent :
func (c *Chaussette) WaitEvent(events *msg.Iterator, topicName string, eventName string, timeout int) *msg.Event {
	term := make(chan *msg.Event, 1)
	cont := true
	go func() {
		for cont {
			message := events.Get()
			if message != nil {
				event := (*message).(msg.Event)
				if topicName == event.GetTopic() && eventName == event.GetEvent() {
					term <- &event
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

// WaitCommand : uniquement au sein d'un connecteur a priori
func (c *Chaussette) WaitCommand(commands *msg.Iterator, commandName string, timeout int) *msg.Command {
	term := make(chan *msg.Command, 1)
	cont := true
	go func() {
		for cont {
			message := commands.Get()
			if message != nil {
				command := (*message).(msg.Command)
				if commandName == command.GetCommand() {
					term <- &command
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

// WaitReply :
func (c *Chaussette) WaitReply(replies *msg.Iterator, commandUUID string, timeout int) *msg.Reply {
	term := make(chan *msg.Reply, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				reply := (*message).(msg.Reply)
				if commandUUID == reply.GetCmdUUID() {
					term <- &reply
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

// WaitConfig :
func (c *Chaussette) WaitConfig(replies *msg.Iterator, commandUUID string, timeout int) *msg.Config {
	term := make(chan *msg.Config, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				config := (*message).(msg.Config)
				term <- &config
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
			}
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}

func chaussetteClient(logicalName string, address string) {
	c := NewChaussette(logicalName)
	c.Connect(address)
	time.Sleep(time.Second * time.Duration(1))
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(5))
			config := c.NewHandshakeMessage()
			fmt.Println(config.String())
		}
	}()
	/*
		go func() {
			command := msg.NewCommand("orchestrator", "deploy", "{\"appli\": \"toto\"}")
			c.SendCommand(command)
			event := msg.NewEvent("bus", "coucou", "ok")
			c.SendEvent(event)

			events := msg.NewIterator(c.qEvents)
			defer events.Close()
			rec := c.WaitEvent(events, "bus", "started", 20)
			if rec != nil {
				fmt.Printf(">Received Event: \n%#v\n", *rec)
			} else {
				fmt.Print("Timeout expired !")
			}
			events2 := msg.NewIterator(c.qEvents)
			defer events.Close()
			rec2 := c.WaitEvent(events2, "bus", "starting", 20)
			if rec2 != nil {
				fmt.Printf(">Received Event 2: \n%#v\n", *rec2)
			} else {
				fmt.Print("Timeout expired  2 !")
			}
		}()

	*/
	<-c.done
}

func chaussetteServer(logicalName string, address string) {
	s := NewChaussette(logicalName)
	err := s.Bind(address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	}
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(5))
			config := s.NewHandshakeMessage()
			fmt.Println(config.String())
		}
	}()
	/*
		go func() {
			time.Sleep(time.Second * time.Duration(5))
			event := msg.NewEvent("bus", "starting", "ok")
			s.SendEvent(event)
			time.Sleep(time.Millisecond * time.Duration(200))
			event = msg.NewEvent("bus", "started", "ok")
			s.SendEvent(event)
			command := msg.NewCommand("bus", "register", "{\"topic\": \"toto\"}")
			s.SendCommand(command)
			reply := msg.NewReply(command, "success", "OK")
			s.SendReply(reply)
		}()
	*/
	<-s.done
}

func chaussetteTest() {
	done := make(chan bool)

	fmt.Printf("\n--\ncreation c1\n")
	c1 := NewChaussette("c")
	c1.Bind("localhost:8301")

	fmt.Printf("\n--\ncreation c2\n")
	c2 := NewChaussette("c")
	c2.Bind("localhost:8302")

	fmt.Printf("\n--\ncreation b1\n")
	b1 := NewChaussette("b")
	b1.Bind("localhost:8201")
	b1.Connect("localhost:8302")
	b1.Connect("localhost:8301")

	fmt.Printf("\n--\ncreation a1\n")
	a1 := NewChaussette("a")
	a1.Bind("localhost:8001")
	a1.Connect("localhost:8201")

	fmt.Printf("\n--\ncreation b2\n")
	b2 := NewChaussette("b")
	b2.Bind("localhost:8202")
	b2.Connect("localhost:8301")

	time.Sleep(time.Second * time.Duration(1))
	fmt.Printf("\n--\nb1 config b1\n  %s\n", b1.NewHandshakeMessage().String())
	fmt.Printf("\n--\nb2 config b2\n  %s\n", b2.NewHandshakeMessage().String())

	<-done
}
