package shoset

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	//	uuid "github.com/kjk/betterguid"

	"github.com/ditrit/shoset/msg"
)

// ShosetConn : client connection
type ShosetConn struct {
	socket           *tls.Conn
	remoteLname      string // logical name of the socket in fornt of this one
	remoteShosetType string // shosetType of the socket in fornt of this one
	dir              string
	remoteAddress    string // addresse of the socket in fornt of this one
	ch               *Shoset
	rb               *msg.Reader
	wb               *msg.Writer
	isValid          bool // for join protocol
}

// GetDir :
func (c *ShosetConn) GetDir() string { return c.dir }

// GetCh :
func (c *ShosetConn) GetCh() *Shoset { return c.ch }

func (c *ShosetConn) GetLocalLogicalName() string { return c.ch.GetLogicalName() }

// GetName : // remote logical Name
func (c *ShosetConn) GetRemoteLogicalName() string { return c.remoteLname }

// GetShosetType : // remote ShosetTypeName
func (c *ShosetConn) GetRemoteShosetType() string { return c.remoteShosetType }

// GetBindAddr : port sur lequel on est bindé
func (c *ShosetConn) GetLocalAddress() string { return c.ch.GetBindAddress() }

func (c *ShosetConn) GetRemoteAddress() string { return c.remoteAddress }

func (c *ShosetConn) GetIsValid() bool { return c.isValid }

// SetName : // remote logical Name
func (c *ShosetConn) SetRemoteLogicalName(lName string) { // remote logical Name
	c.remoteLname = lName // remote logical Name
	// c.GetCh().ConnsByName.Set(c.GetName(), c.GetRemoteAddress(), c)
}

// SetBindAddr :
func (c *ShosetConn) SetLocalAddress(bindAddress string) {
	if bindAddress != "" {
		c.ch.SetBindAddress(bindAddress)
	}
}

// SetShosetType : // remote ShosetType
func (c *ShosetConn) SetShosetType(ShosetType string) {
	if ShosetType != "" {
		c.remoteShosetType = ShosetType
	}
}

func (c *ShosetConn) SetIsValid(state bool) {
	c.isValid = state
}

func (c *ShosetConn) SetRemoteAddress(address string) {
	if address != "" {
		c.remoteAddress = address
	}
}

func NewShosetConn(c *Shoset, address string, dir string) (*ShosetConn, error) {
	// Creation
	conn := ShosetConn{}
	// Initialisation attributs ShosetConn
	conn.ch = c
	conn.dir = dir
	conn.socket = new(tls.Conn)
	conn.rb = new(msg.Reader)
	conn.wb = new(msg.Writer)
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}
	conn.remoteAddress = ipAddress
	conn.isValid = true
	return &conn, nil
}

func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{ name : %s, way : %s, remoteAddress : %s}", c.GetRemoteLogicalName(), c.GetDir(), c.GetRemoteAddress())
}

// ReadString :
func (c *ShosetConn) ReadString() (string, error) {
	fmt.Println("enter readstring ~~~~")
	return c.rb.ReadString()
}

// ReadMessage :
func (c *ShosetConn) ReadMessage(data interface{}) error {
	return c.rb.ReadMessage(data)
}

// WriteString :
func (c *ShosetConn) WriteString(data string) (int, error) {
	return c.wb.WriteString(data)
}

// Flush :
func (c *ShosetConn) Flush() error {
	return c.wb.Flush()
}

// WriteMessage :
func (c *ShosetConn) WriteMessage(data interface{}) error {
	return c.wb.WriteMessage(data)
}

// RunOutConn : handler for the socket, for Link()
func (c *ShosetConn) runOutConn() {
	// fmt.Println("Entering runoutconn")
	// fmt.Println("c.ch.lname = ", c.ch.lName)
	myConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "link")
	for {
		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig)
		defer conn.Close()
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		} else {
			// fmt.Println("!!!!!!!!!!!!! init socket conn, name : ", c.GetName())
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)

			// receive messages
			for {
				// fmt.Println("enter for loop runoutconn")
				if c.GetRemoteLogicalName() == "" { // remote logical Name // same problem than runJoinConn()
					c.SendMessage(*myConfig)
				}
				// c.SendMessage(*myConfig)
				err := c.receiveMsg("link")
				time.Sleep(time.Second * time.Duration(1))
				if err != nil {
					// fmt.Println("error detected in receiving msg 2")
					c.SetRemoteLogicalName("") // reinitialize conn
					break
				}
			}
		}
	}
}

// RunJoinConn : handler for the socket, for Join()
func (c *ShosetConn) runJoinConn() {
	// fmt.Println("########### enter runjoinconn")
	joinConfig := msg.NewCfg(c.ch.GetBindAddress(), c.ch.GetLogicalName(), c.ch.GetShosetType(), "join") //we create a new message config
	for {
		// fmt.Println("in for loop from runjoinconn")
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name
			// fmt.Println("c is not valid")
			break
		}
		// 	ch.ConnsJoin.Set(c.addr, c)       // à déplacer une fois
		// 	ch.NameBrothers.Set(c.addr, true) // les connexions établies (fin de fonction)

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		} else { // a connection occured
			// fmt.Printf("\n########### a connection occured")
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			defer conn.Close()

			// receive messages
			for {
				// fmt.Println("~~~~", ch.ConnsJoin.Get(c.GetBindAddr()))
				// fmt.Println("\n########### enter receive message loop")
				// fmt.Println("connsJoin : ", ch.ConnsByName.Get(ch.GetName()))
				// if connsJoin := ch.ConnsByName.Get(ch.GetLogicalName()); connsJoin != nil {
				// 	if exists := connsJoin.Get(c.GetLocalAddress()); exists == nil {
				// 		c.SendMessage(*joinConfig)
				// 	}
				// }
				if c.GetRemoteLogicalName() == "" { // remote logical Name // same problem than runJoinConn()
					c.SendMessage(*joinConfig)
				}
				// c.SendMessage(*joinConfig)

				// fmt.Println(c.GetLocalAddress(), " can receive message from ", c.GetRemoteAddress())
				err := c.receiveMsg("join")
				// fmt.Println(c.GetIsValid(), " - after message received - in runjoinconn")
				time.Sleep(time.Second * time.Duration(1))
				if err != nil {
					// fmt.Println("error detected in receiving msg")
					c.SetRemoteLogicalName("") // reinitialize conn
					break
				}
			}
		}
	}
}

// runConn : handler for the socket, for Protocol()
func (c *ShosetConn) runConn(config *msg.ConfigProtocol, protocolType string) {
	fmt.Println("remote address : ", c.GetRemoteAddress())
	// time.Sleep(time.Second * time.Duration(1))
	// fmt.Println("tlsconfig : ", c.ch.tlsConfig)
	for {
		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		} else {
			// fmt.Println("enter runconn!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			defer conn.Close()

			for {
				if c.GetRemoteLogicalName() == "" {
					c.SendMessage(*config)
				}

				err := c.receiveMsg(protocolType)
				time.Sleep(time.Second * time.Duration(1))
				if err != nil {
					c.SetRemoteLogicalName("") // reinitialize conn
					break
				}
			}
		}
	}
}

// runInConn : handler for the connection, for handleBind()
func (c *ShosetConn) runInConn(protocolType string) {
	// fmt.Println("enter runinconn")
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)
	defer c.socket.Close()
	// receive messages
	for {
		err := c.receiveMsg(protocolType)
		time.Sleep(time.Millisecond * time.Duration(10))
		if err != nil {
			return
		}
	}
}

// SendMessage :
func (c *ShosetConn) SendMessage(msg msg.Message) {
	//fmt.Printf("     Sending message %s(%s) -> %s(%s) %#v.\n", c.GetCh().GetName(), c.GetCh().GetBindAddr(), c.GetName(), c.addr, msg)
	// fmt.Println("\nenter send message")
	c.WriteString(msg.GetMsgType())
	c.WriteMessage(msg)
}

func (c *ShosetConn) receiveMsg(protocolType string) error {
	// fmt.Println("enter receive message ", c.GetLocalAddress())
	if !c.GetIsValid() {
		// fmt.Println("c is not valid !!!!!!!!", c.GetLocalAddress())
		c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName(), protocolType)
		return errors.New("error : Invalid connection for join - not the same type/name")
	}
	// fmt.Println(c.ch.bindAddress)
	// fmt.Printf("\n########### receiveMsg : gonna read message")
	// read message type
	msgType, err := c.rb.ReadString() /////////////////////////////////////////////////////////////// problem here : doesn't exit RedaString() - fixed by changing sth in runJoinConn()
	// fmt.Printf("\n########### receiveMsg : message read")
	switch {
	case err == io.EOF:
		// fmt.Println("\n########### receiveMsg : reached EOF - close this connection", c.GetRemoteAddress())
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName(), protocolType)
		}
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		// fmt.Println("\n########### receiveMsg : failed to read - close this connection")
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName(), protocolType)
		}
		return errors.New("error : receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")
	// fmt.Println("\n########### receiveMsg : ", msgType)
	// read Message Value
	fGet, ok := c.ch.Get[msgType]
	if ok {
		// fmt.Println("\n########### receiveMsg : message ok")
		msgVal, err := fGet(c)
		if err == nil {
			// read message data and handle it with the proper function
			fHandle, ok := c.ch.Handle[msgType]
			if ok {
				go fHandle(c, msgVal) //HandleConfigJoin() or HandleConfigLink()
			}
		} else {
			if c.GetDir() == "in" {
				c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName(), protocolType)
			}
			// fmt.Println("receiveMsg : can not read value of " + msgType)
			return errors.New("receiveMsg : can not read value of " + msgType)
		}
	}
	if !ok {
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName(), protocolType)
		}
		// fmt.Println("receiveMsg : non implemented type of message " + msgType)
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	time.Sleep(time.Second * time.Duration(1))
	// fmt.Println(c.GetIsValid(), " - after receivemsg")
	return nil
}
