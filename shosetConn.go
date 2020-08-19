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
}

func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{ way: %s, lName: %s, Type: %s, addr(bindAddr): %s(%s) }", c.dir, c.name, c.ShosetType, c.addr, c.bindAddr)
}

// ReadString :
func (c *ShosetConn) ReadString() (string, error) {
	return c.rb.ReadString()
}

// ReadMessage :
func (c *ShosetConn) ReadMessage(data interface{}) error {
	fmt.Printf("readmessage shosetconn 1\n")
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

// runInbound : handler for the connection
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

// RunOutConn : handler for the socket
func (c *ShosetConn) runOutConn(addr string) {
	ch := c.GetCh()
	myConfig := NewHandshake(ch)
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
			fmt.Printf("connection successful\n")

			errSend := c.SendMessage(*myConfig)
			if errSend != nil {
				fmt.Println(errSend.Error())
			}

			// receive messages
			for {
				errRec := c.receiveMsg()
				if errRec != nil {
					fmt.Println(errRec.Error())
					break
				}
			}
		}
	}
}

// RunJoinConn : handler for the socket
func (c *ShosetConn) runJoinConn() {
	ch := c.GetCh()
	joinConfig := msg.NewCfgJoin(ch.GetBindAddr())
	for {
		fmt.Printf("runJoinConn start connection on %s\n", ch.bindAddr)
		ch.ConnsJoin.Set(c.addr, c)
		ch.NameBrothers.Set(c.addr, true)
		conn, errConn := tls.Dial("tcp", c.addr, ch.tlsConfig)
		defer conn.Close()
		if errConn != nil {
			time.Sleep(time.Second * time.Duration(5))
		} else {
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)
			c.name = ""
			fmt.Printf("connection successful\n")
			// receive messages
			for {
				fmt.Printf("receive loop starts\n")
				if c.name == "" { // remote logical Name
					errSend := c.SendMessage(*joinConfig)
					if errSend != nil {
						fmt.Println(errSend.Error())
						break
					}
				}
				errRec := c.receiveMsg()
				if errRec != nil {
					fmt.Println(errRec.Error())
					break
				}
			}
		}
	}
}

// GetDir :
func (c *ShosetConn) GetDir() string {
	return c.dir
}

// GetCh :
func (c *ShosetConn) GetCh() *Shoset {
	return c.ch
}

// GetName : // remote logical Name
func (c *ShosetConn) GetName() string { // remote logical Name
	return c.name // remote logical Name
}

// GetShosetType : // remote ShosetTypeName
func (c *ShosetConn) GetShosetType() string { return c.ShosetType }

// GetBindAddr :
func (c *ShosetConn) GetBindAddr() string {
	return c.bindAddr
}

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

// SendMessage :
func (c *ShosetConn) SendMessage(msg msg.Message) error {
	//fmt.Printf("     Sending message %s(%s) -> %s(%s) %#v.\n", c.GetCh().GetName(), c.GetCh().GetBindAddr(), c.GetName(), c.addr, msg)
	c.WriteString(msg.GetMsgType())
	return c.WriteMessage(msg)
}

// receiveMsg
func (c *ShosetConn) receiveMsg() error {
	// read message type
	msgType, err := c.rb.ReadString()
	switch {
	case err == io.EOF:
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")
	// read Message Value
	fGet, ok := c.ch.Get[msgType]
	fmt.Printf("message type : %v, sent by %v\n", msgType, c.bindAddr)
	if ok {
		fmt.Printf(" ok\n")
		msgVal, err := fGet(c)
		fmt.Printf(" - err : %v\n", err)
		if err == nil {
			// read message data and handle it
			fHandle, ok := c.ch.Handle[msgType]
			if ok {
				fmt.Printf("  before handle\n")
				go fHandle(c, msgVal)
			}
		} else {
			c.ch.deleteConn(c.addr)
			return errors.New("receiveMsg : can not read value of " + msgType)
		}
	}
	if !ok {
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	return nil
}
