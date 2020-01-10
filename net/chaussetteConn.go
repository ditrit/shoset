package net

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	//	uuid "github.com/kjk/betterguid"

	"../msg"
)

// ChaussetteConn : client connection
type ChaussetteConn struct {
	socket   *tls.Conn
	name     string // remote logical name
	bindAddr string // remote bind addr
	dir      string
	addr     string
	ch       *Chaussette
	rb       *msg.Reader
	wb       *msg.Writer
}

func (c *ChaussetteConn) String() string {
	return fmt.Sprintf("ChaussetteConn{ way: %s, lName: %s, addr(bindAddr): %s(%s)", c.dir, c.name, c.addr, c.bindAddr)
}

// ReadString :
func (c *ChaussetteConn) ReadString() (string, error) {
	return c.rb.ReadString()
}

// ReadMessage :
func (c *ChaussetteConn) ReadMessage(data interface{}) error {
	return c.rb.ReadMessage(data)
}

// WriteString :
func (c *ChaussetteConn) WriteString(data string) (int, error) {
	return c.wb.WriteString(data)
}

// Flush :
func (c *ChaussetteConn) Flush() error {
	return c.wb.Flush()
}

// WriteMessage :
func (c *ChaussetteConn) WriteMessage(data interface{}) error {
	return c.wb.WriteMessage(data)
}

// RunOutConn : handler for the socket
func (c *ChaussetteConn) runOutConn(addr string) {
	myConfig := msg.NewHandshake(c.ch.bindAddr, c.ch.lName)
	for {
		c.ch.SetConn(addr, c)
		conn, err := tls.Dial("tcp", c.addr, c.ch.tlsConfig)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
		} else {
			c.socket = conn
			c.rb = msg.NewReader(c.socket)
			c.wb = msg.NewWriter(c.socket)

			// receive messages
			for {
				if c.name == "" { // remote logical Name
					c.ch.sendConn["cfg"](c, myConfig)
				}
				c.receiveMsg()
			}
		}
	}
}

// GetDir :
func (c *ChaussetteConn) GetDir() string {
	return c.dir
}

// GetCh :
func (c *ChaussetteConn) GetCh() *Chaussette {
	return c.ch
}

// GetName : // remote logical Name
func (c *ChaussetteConn) GetName() string { // remote logical Name
	return c.name // remote logical Name
}

// GetBindAddr :
func (c *ChaussetteConn) GetBindAddr() string {
	return c.bindAddr
}

// SetName : // remote logical Name
func (c *ChaussetteConn) SetName(lName string) { // remote logical Name
	if lName != "" {
		c.name = lName // remote logical Name
		if c.ch.connsByName[lName] == nil {
			c.ch.connsByName[lName] = make(map[string]*ChaussetteConn)
		}
		c.ch.connsByName[lName][c.addr] = c
	}
}

// SetBindAddr :
func (c *ChaussetteConn) SetBindAddr(bindAddr string) {
	if bindAddr != "" {
		c.bindAddr = bindAddr
	}
}

// runInbound : handler for the connection
func (c *ChaussetteConn) runInConn() {
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)
	myConfig := msg.NewHandshake(c.ch.bindAddr, c.ch.lName)

	// receive messages
	for {
		if c.name == "" { // remote logical Name
			c.ch.sendConn["cfg"](c, myConfig)
		}
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
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : reached EOF - close this connection")
	case err != nil:
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : failed to read - close this connection")
	}
	msgType = strings.Trim(msgType, "\n")

	// read message data and handle it
	fhandle, ok := c.ch.handle[msgType]
	if ok {
		fhandle(c)
	} else {
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	if err != nil {
		c.ch.deleteConn(c.addr)
		return errors.New("receiveMsg : unable to decode a message of type  " + msgType)
	}
	return err
}
