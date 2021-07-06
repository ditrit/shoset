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
	socket        *tls.Conn
	name          string // remote logical name
	ShosetType    string // remote ShosetType
	brothers      map[string]bool
	dir           string
	remoteAddress string // addresse de la chaussette en face
	ch            *Shoset
	rb            *msg.Reader
	wb            *msg.Writer
	isValid       bool
}

// GetDir :
func (c *ShosetConn) GetDir() string { return c.dir }

// GetCh :
func (c *ShosetConn) GetCh() *Shoset { return c.ch }

// GetName : // remote logical Name
func (c *ShosetConn) GetName() string { return c.name }

// GetShosetType : // remote ShosetTypeName
func (c *ShosetConn) GetShosetType() string { return c.ShosetType }

// GetBindAddr : port sur lequel on est bindé
func (c *ShosetConn) GetLocalAddress() string { return c.ch.GetBindAddress() }

func (c *ShosetConn) GetRemoteAddress() string { return c.remoteAddress }

func (c *ShosetConn) GetIsValid() bool { return c.isValid }

// SetName : // remote logical Name
func (c *ShosetConn) SetName(lName string) { // remote logical Name
	if lName != "" {
		c.name = lName // remote logical Name
		c.GetCh().ConnsByName.Set(c.GetName(), c.GetRemoteAddress(), c)
	}
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
		c.ShosetType = ShosetType
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
	// conn.bindAddr = ipAddress // delete
	conn.brothers = make(map[string]bool)
	conn.isValid = true
	return &conn, nil
}

func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{ way: %s, remoteAddress : %s}", c.GetDir(), c.GetRemoteAddress())
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
func (c *ShosetConn) runOutConn(addr string) {
	// fmt.Println("Entering runoutconn")
	myConfig := NewHandshake(c.GetCh())
	for {
		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig)
		defer conn.Close()
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			c.ch.SetConn(addr, c.ch.GetShosetType(), c)

			// receive messages
			for {
				if c.GetName() == "" { // remote logical Name // same problem than runJoinConn()
					c.SendMessage(*myConfig)
				}
				c.receiveMsg()
			}
		}
	}
}

// RunJoinConn : handler for the socket, for Join()
func (c *ShosetConn) runJoinConn() {
	// fmt.Println("########### enter runjoinconn")
	ch := c.GetCh()
	// fmt.Println(ch.GetBindAddr())                                                        // socket from socketConn
	joinConfig := msg.NewCfg(ch.GetBindAddress(), ch.GetName(), ch.GetShosetType(), "join") //we create a new message config
	for {
		// fmt.Println("in for loop from runjoinconn")
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name
			// fmt.Println("c is not valid")
			break
		}
		// 	ch.ConnsJoin.Set(c.addr, c)       // à déplacer une fois
		// 	ch.NameBrothers.Set(c.addr, true) // les connexions établies (fin de fonction)

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), ch.tlsConfig) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Second * time.Duration(1))
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
				if connsJoin := ch.ConnsByName.Get(ch.GetName()); connsJoin != nil {
					if exists := connsJoin.Get(c.GetLocalAddress()); exists == nil {
						c.SendMessage(*joinConfig)
					}
				}
				c.SendMessage(*joinConfig)

				// fmt.Println(c.GetLocalAddress(), " can receive message from ", c.GetRemoteAddress())
				err := c.receiveMsg()
				// fmt.Println(c.GetIsValid(), " - after message received - in runjoinconn")
				time.Sleep(time.Second * time.Duration(1))
				if err != nil {
					// fmt.Println("error detected in receiving msg")
					break
				}
			}
		}
	}
}

// runInbound : handler for the connection, for handleBind()
func (c *ShosetConn) runInConn() {
	// fmt.Println("enter runinconn")
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)
	defer c.socket.Close()
	// receive messages
	for {
		err := c.receiveMsg()
		time.Sleep(time.Millisecond * time.Duration(10))
		if err != nil {
			return
		}
	}
}

// SendMessage :
func (c *ShosetConn) SendMessage(msg msg.Message) {
	//fmt.Printf("     Sending message %s(%s) -> %s(%s) %#v.\n", c.GetCh().GetName(), c.GetCh().GetBindAddr(), c.GetName(), c.addr, msg)
	// fmt.Printf("\n########### enter send message")
	c.WriteString(msg.GetMsgType())
	c.WriteMessage(msg)
}

func (c *ShosetConn) receiveMsg() error {
	// fmt.Println("###########! enter receive message ", c.GetLocalAddress())
	if !c.GetIsValid() {
		// fmt.Println("c is not valid !!!!!!!!", c.GetLocalAddress())
		c.ch.deleteConn(c.GetRemoteAddress())
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
			c.ch.deleteConn(c.GetRemoteAddress())
		}
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		// fmt.Println("\n########### receiveMsg : failed to read - close this connection")
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress())
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
				c.ch.deleteConn(c.GetRemoteAddress())
			}
			// fmt.Println("receiveMsg : can not read value of " + msgType)
			return errors.New("receiveMsg : can not read value of " + msgType)
		}
	}
	if !ok {
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress())
		}
		// fmt.Println("receiveMsg : non implemented type of message " + msgType)
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	time.Sleep(time.Second * time.Duration(1))
	// fmt.Println(c.GetIsValid(), " - after receivemsg")
	return nil
}
