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
	socket     *tls.Conn
	name       string // remote logical name
	ShosetType string // remote ShosetType
	bindAddr   string // remote bind addr
	brothers   map[string]bool
	dir        string
	addr       string
	ch         *Shoset
	rb         *msg.Reader
	wb         *msg.Writer
	isValid    bool
}

// GetDir :
func (c *ShosetConn) GetDir() string { return c.dir }

// GetCh :
func (c *ShosetConn) GetCh() *Shoset { return c.ch }

// GetName : // remote logical Name
func (c *ShosetConn) GetName() string { return c.name }

// GetShosetType : // remote ShosetTypeName
func (c *ShosetConn) GetShosetType() string { return c.ShosetType }

// GetBindAddr :
func (c *ShosetConn) GetBindAddr() string { return c.bindAddr }

// SetName : // remote logical Name
func (c *ShosetConn) SetName(lName string) { // remote logical Name
	if lName != "" {
		c.name = lName // remote logical Name
		c.GetCh().ConnsByName.Set(c.name, c.addr, c)
	}
}

// SetShosetType : // remote ShosetType
func (c *ShosetConn) SetShosetType(ShosetType string) {
	if ShosetType != "" {
		c.ShosetType = ShosetType
	}
}

// SetBindAddr :
func (c *ShosetConn) SetBindAddr(bindAddr string) {
	if bindAddr != "" {
		c.bindAddr = bindAddr
	}
}

func (c *ShosetConn) SetIsValid(state bool) {
	c.isValid = state
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
	conn.addr = ipAddress
	conn.bindAddr = ipAddress
	conn.brothers = make(map[string]bool)
	conn.isValid = true
	return &conn, nil
}

func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{ way: %s, lName: %s, Type: %s, addr(bindAddr): %s(%s)", c.dir, c.name, c.ShosetType, c.addr, c.bindAddr)
}

// ReadString :
func (c *ShosetConn) ReadString() (string, error) {
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
	myConfig := NewHandshake(c.GetCh())
	for {
		conn, err := tls.Dial("tcp", c.addr, c.ch.tlsConfig)
		defer conn.Close()
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			c.ch.SetConn(addr, c.ShosetType, c)

			// receive messages
			for {
				if c.name == "" { // remote logical Name // same problem than runJoinConn()
					c.SendMessage(*myConfig)
				}
				c.receiveMsg()
			}
		}
	}
}

// RunJoinConn : handler for the socket, for Join()
func (c *ShosetConn) runJoinConn() {
	// fmt.Printf("########### enter runjoinconn")
	ch := c.GetCh()                                                                  // socket from socketConn
	joinConfig := msg.NewCfgJoin(ch.GetBindAddr(), ch.GetName(), ch.GetShosetType()) //we create a new message config
	for {
		if !c.isValid { // sockets are not from the same type or don't have the same name
			fmt.Println("error : Invalid connection for join - not the same type/name")
			break
		}
		// 	ch.ConnsJoin.Set(c.addr, c)       // à déplacer une fois
		// 	ch.NameBrothers.Set(c.addr, true) // les connexions établies (fin de fonction)

		conn, err := tls.Dial("tcp", c.addr, ch.tlsConfig) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Second * time.Duration(1))
		} else { // a connection occured
			// fmt.Printf("\n########### a connection occured")
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			defer conn.Close()

			// receive messages
			for {
				// fmt.Printf("\n########### enter receive message loop")
				if ch.ConnsJoin.Get(c.GetBindAddr()) == nil { ///////////////////////////////////////////////////////////////// problem here with != : doesn't enter if - fixed with ==
					// fmt.Printf("\n########### can send message")
					c.SendMessage(*joinConfig)
				}
				// fmt.Printf("\n########### can receive message")
				err := c.receiveMsg()
				if err != nil {
					if err == io.EOF {
						fmt.Println("receiveMsg : reached EOF - close this connection")
					} else {
						fmt.Println("receiveMsg : failed to read - close this connection")
					}
					break
				}
			}
		}
	}
}

// runInbound : handler for the connection, for handleBind()
func (c *ShosetConn) runInConn() {
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
	// fmt.Printf("\n########### enter receive message")
	if !c.isValid {
		c.ch.deleteConn(c.addr)
		fmt.Printf("error : Invalid connection for join - not the same type/name")
		return errors.New("error : Invalid connection for join - not the same type/name")
	}
	// fmt.Printf("\n########### receiveMsg : gonna read message")
	// read message type
	msgType, err := c.rb.ReadString() /////////////////////////////////////////////////////////////// problem here : doesn't exit RedaString() - fixed by changing sth in runJoinConn()
	// fmt.Printf("\n########### receiveMsg : message read")
	switch {
	case err == io.EOF:
		fmt.Printf("\n########### receiveMsg : reached EOF - close this connection")
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		fmt.Printf("\n########### receiveMsg : failed to read - close this connection")
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")
	// fmt.Printf("\n########### receiveMsg : no err")
	// read Message Value
	fGet, ok := c.ch.Get[msgType]
	if ok {
		// fmt.Printf("\n########### receiveMsg : message ok")
		msgVal, err := fGet(c)
		if err == nil {
			// read message data and handle it
			fHandle, ok := c.ch.Handle[msgType]
			if ok {
				go fHandle(c, msgVal) //HandleConfigJoin()
			}
		} else {
			c.ch.deleteConn(c.addr)
			fmt.Printf("receiveMsg : can not read value of " + msgType)
			return errors.New("receiveMsg : can not read value of " + msgType)
		}
	}
	if !ok {
		c.ch.deleteConn(c.addr)
		fmt.Printf("receiveMsg : non implemented type of message " + msgType)
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	return nil
}
