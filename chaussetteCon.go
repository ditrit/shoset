package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	//	uuid "github.com/kjk/betterguid"

	"./msg"
)

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

func (c *ChaussetteConn) String() string {
	return fmt.Sprintf("ChaussetteConn{ way: %s, logicalName: %s, address(bindAddress): %s(%s)", c.direction, c.remoteLogicalName, c.address, c.remoteBindAddress)
}

// RunOutConn : handler for the socket
func (c *ChaussetteConn) runOutConn(address string) {
	myConfig := msg.NewHandshake(c.chaussette.bindAddress, c.chaussette.logicalName)
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

// runInbound : handler for the connection
func (c *ChaussetteConn) runInConn() {
	c.rb = msg.NewReader(c.socket)
	c.wb = msg.NewWriter(c.socket)
	myConfig := msg.NewHandshake(c.chaussette.bindAddress, c.chaussette.logicalName)

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
				newBrother := []string{c.remoteBindAddress}
				cfgNewBrother := msg.NewBrothers(newBrother)
				oldBrothers := []string{}
				cfgNewInstance := c.chaussette.NewInstanceMessage(c.remoteBindAddress, c.remoteLogicalName)
				fmt.Printf("\n newInstance : %#v\n", cfgNewInstance)
				for _, conn := range c.chaussette.connsByName[c.remoteLogicalName] {
					if conn.direction == "in" && conn != c {
						conn.SendConfig(cfgNewInstance)
						oldBrothers = append(oldBrothers, conn.remoteBindAddress)
						fmt.Printf("    Sending NewBrother %s to %s:%s\n", c.remoteBindAddress, c.remoteLogicalName, conn.remoteBindAddress)
						conn.SendConfig(cfgNewBrother)
					}
				}
				cfgOldBrothers := msg.NewBrothers(oldBrothers)
				fmt.Printf("  Sending oldBrothers %#v to %s:%s \n", oldBrothers, c.remoteLogicalName, c.remoteBindAddress)
				c.SendConfig(cfgOldBrothers)

				for connAddr := range c.chaussette.brothers {
					cfgConnectTo := c.chaussette.NewConnectToMessage(connAddr)
					c.SendConfig(cfgConnectTo)
				}
			}
		}
		if c.direction == "out" {
			fmt.Printf("\n--\nhandhsake for %s to %s\n", c.chaussette.bindAddress, c.remoteBindAddress)
			newBrother := []string{c.remoteBindAddress}
			cfgNewBrother := msg.NewBrothers(newBrother)
			oldBrothers := []string{}
			for _, conn := range c.chaussette.connsByName[c.remoteLogicalName] {
				if conn.direction == "out" && conn != c {
					oldBrothers = append(oldBrothers, conn.remoteBindAddress)
					fmt.Printf("    Sending NewBrother %s to %s\n", c.remoteBindAddress, conn.remoteBindAddress)
					conn.SendConfig(cfgNewBrother)
				}
			}
			fmt.Printf("  --> oldBrothers : %#v\n", oldBrothers)
			cfgOldBrothers := msg.NewBrothers(oldBrothers)
			fmt.Printf("  Sending oldBrothers to %s\n", c.remoteLogicalName)
			c.SendConfig(cfgOldBrothers)
		}
	case "brothers":
		for _, addr := range cfg.Brothers {
			c.chaussette.brothers[addr] = true
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
