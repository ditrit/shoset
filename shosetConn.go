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

func (c *ShosetConn) GetLocalShosetType() string { return c.ch.GetShosetType() }

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
func (c *ShosetConn) SetRemoteShosetType(ShosetType string) {
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
	return fmt.Sprintf("ShosetConn{ name : %s, type : %s, way : %s, remoteAddress : %s}", c.GetRemoteLogicalName(), c.GetRemoteShosetType(), c.GetDir(), c.GetRemoteAddress())
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
func (c *ShosetConn) runOutConn() {
	myConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "link")
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		} else {
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			defer conn.Close()

			// receive messages
			for {
				if c.GetRemoteLogicalName() == "" {
					c.SendMessage(*myConfig)
				}

				err := c.receiveMsg()
				time.Sleep(time.Millisecond * time.Duration(100))
				if err != nil {
					c.SetRemoteLogicalName("") // reinitialize conn
					break
				}
			}
		}
	}
}

// RunJoinConn : handler for the socket, for Join()
func (c *ShosetConn) runJoinConn() {
	joinConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "join") //we create a new message config
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		} else { // a connection occured
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			defer conn.Close()

			// receive messages
			for {
				if c.GetRemoteLogicalName() == "" {
					c.SendMessage(*joinConfig)
				}

				err := c.receiveMsg()
				time.Sleep(time.Millisecond * time.Duration(100))
				if err != nil {
					c.SetRemoteLogicalName("") // reinitialize conn
					break
				}
			}
		}
	}
}

// runEndConn : handler for the socket, for Bye()
func (c *ShosetConn) runEndConn() {
	// fmt.Println(c.ch.GetBindAddress(), "enter run endconn")
	byeConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "bye") //we create a new message config
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		// fmt.Println(c.ch.GetBindAddress(), "in run endconn")
		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfig) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		} else { // a connection occured
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			defer conn.Close()

			// receive messages
			for {
				if c.GetRemoteLogicalName() == "" {
					c.SendMessage(*byeConfig)
				}

				err := c.receiveMsg()
				time.Sleep(time.Millisecond * time.Duration(100))
				if err != nil {
					c.SetRemoteLogicalName("") // reinitialize conn
					break
				}
			}
		}
	}
}

// runInConn : handler for the connection, for handleBind()
func (c *ShosetConn) runInConn() {
	fmt.Println(c.ch.GetBindAddress(), "in runinconn")
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)
	defer c.socket.Close()

	// receive messages
	for {
		err := c.receiveMsg()
		time.Sleep(time.Millisecond * time.Duration(10))
		if err != nil {
			if err.Error() == "error : Invalid connection for join - not the same type/name or shosetConn ended" {
				c.ch.SetIsValid(false)
				goto Exit
			}
			break
		}
	}
Exit:
}

// SendMessage :
func (c *ShosetConn) SendMessage(msg msg.Message) {
	c.WriteString(msg.GetMsgType())
	c.WriteMessage(msg)
}

func (c *ShosetConn) receiveMsg() error {
	if !c.GetIsValid() {
		c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		return errors.New("error : Invalid connection for join - not the same type/name or shosetConn ended")
	}

	// read message type
	msgType, err := c.rb.ReadString()
	switch {
	case err == io.EOF:
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("error : receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")
	// read Message Value
	fGet, ok := c.ch.Get[msgType]
	if ok {
		msgVal, err := fGet(c)
		if err == nil {
			// read message data and handle it with the proper function
			fHandle, ok := c.ch.Handle[msgType]
			if ok {
				go fHandle(c, msgVal) //HandleConfigJoin() or HandleConfigLink() or HandleConfigBye()
			}
		} else {
			if c.GetDir() == "in" {
				c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
			}
			return errors.New("receiveMsg : can not read value of " + msgType)
		}
	}
	if !ok {
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	time.Sleep(time.Millisecond * time.Duration(100)) // maybe we can remove this sleep time
	return nil
}
