package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"./msg"
)

//Chaussette : client gandalf Socket
type Chaussette struct {
	conns       map[string]*ChaussetteConn
	name        string
	done        chan bool
	bindAddress string
	m           sync.RWMutex
	qEvents     *msg.Queue
	qCommands   *msg.Queue
	qReplies    *msg.Queue
	qConfigs    *msg.Queue
	tlsConfig   *tls.Config
	tlsServerOK bool
}

var certPath = "./cert.pem"
var keyPath = "./key.pem"

// NewChaussette : constructor
func NewChaussette(name string) *Chaussette {
	c := new(Chaussette)
	c.name = name
	c.conns = make(map[string]*ChaussetteConn)
	c.qEvents = msg.NewQueue()
	c.qCommands = msg.NewQueue()
	c.qReplies = msg.NewQueue()
	c.qConfigs = msg.NewQueue()

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
	socket     *tls.Conn
	localName  string
	remoteName string
	address    string
	chaussette *Chaussette
	rb         *msg.Reader
	wb         *msg.Writer
}

//Connect : Connect to another Chaussette
func (c *Chaussette) Connect(address string) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.chaussette = c
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	conn.address = address
	conn.localName = c.name
	go conn.runOutConn(address)
	return conn, nil
}

// RunOutConn : handler for the socket
func (c *ChaussetteConn) runOutConn(address string) {
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
				c.receiveMsg()
			}
		}
	}
}

func (c *Chaussette) deleteConn(connAddr string) {
	c.m.Lock()
	delete(c.conns, connAddr)
	c.m.Unlock()
}

func (c *Chaussette) setConn(connAddr string, conn *ChaussetteConn) {
	c.m.Lock()
	c.conns[connAddr] = conn
	c.m.Unlock()
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
	c.bindAddress = address
	go c.handleBind()
	return nil
}

// runBindTo : handler for the socket
func (c *Chaussette) handleBind() error {
	listener, err := net.Listen("tcp", c.bindAddress)
	if err != nil {
		fmt.Println("Failed to bind:", err.Error())
		fmt.Print("GSServer initialized\n")
		return err
	}
	defer listener.Close()

	for {
		unencConn, err := listener.Accept()
		if err != nil {
			fmt.Printf("server: accept %s", err)
			break
		}
		tlsConn := tls.Server(unencConn, c.tlsConfig)
		conn, _ := c.inboudConn(tlsConn)
		fmt.Printf("GSServer : accepted from %s", conn.address)
		go conn.runInConn()
	}
	return nil
}

//inboudConn : Add a new connection from a client
func (c *Chaussette) inboudConn(tlsConn *tls.Conn) (*ChaussetteConn, error) {
	conn := new(ChaussetteConn)
	conn.socket = tlsConn
	conn.chaussette = c
	conn.address = tlsConn.RemoteAddr().String()
	conn.localName = c.name
	c.setConn(conn.address, conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	return conn, nil
}

// runInbound : handler for the connection
func (c *ChaussetteConn) runInConn() {
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)

	// receive messages
	for {
		err := c.receiveMsg()
		if err != nil {
			return
		}
	}

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
	fmt.Printf("Read message and push if into buffer")
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
		var cfg msg.Config
		err = c.rb.ReadConfig(&cfg)
		c.chaussette.qConfigs.Push(cfg)
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

// SendEvent : send an event...
// event is sent on each connection
func (c *Chaussette) SendEvent(evt *msg.Event) {
	fmt.Print("Sending event.\n")
	for _, conn := range c.conns {
		conn.wb.WriteString("evt")
		conn.wb.WriteEvent(*evt)
	}
}

// SendCommand : Send a message
// todo : manage routing
//    identify relevant targets (routing info matches identity)
//    then try on each instance until success
func (c *Chaussette) SendCommand(cmd *msg.Command) {
	fmt.Print("Sending command.\n")
	for _, conn := range c.conns {
		conn.wb.WriteString("cmd")
		conn.wb.WriteCommand(*cmd)
	}
}

// SendReply :
func (c *Chaussette) SendReply(rep *msg.Reply) {
	fmt.Print("Sending reply.\n")
	for _, conn := range c.conns {
		conn.wb.WriteString("rep")
		conn.wb.WriteReply(*rep)
	}
}

// SendConfig :
func (c *Chaussette) SendConfig() {
	fmt.Print("Sending configuration.\n")

	cfg := msg.NewConfig(c.name)
	for _, conn := range c.conns {
		conn.wb.WriteString("evt")
		conn.wb.WriteConfig(*cfg)
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

func chaussetteClient(name string, address string) {
	c := NewChaussette(name)
	c.Connect(address)
	time.Sleep(time.Second * time.Duration(1))
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

	<-c.done
}

func chaussetteServer(name string, address string) {
	s := NewChaussette(name)
	err := s.Bind(address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	}
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
	<-s.done
}
