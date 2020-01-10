package net

import (
	"fmt"

	"../msg"
)

// HandleConfig :
func HandleConfig(c *ChaussetteConn) error {
	var cfg msg.Config
	err := c.ReadMessage(&cfg)
	switch cfg.GetCommandName() {
	case "handshake":
		if c.GetName() == "" {
			c.SetName(cfg.GetLogicalName())
		}
		if c.GetBindAddr() == "" {
			c.SetBindAddr(cfg.GetBindAddress())
		}
		if c.GetDir() == "in" {
			if c.GetBindAddr() != "" {
				newBrother := []string{c.GetBindAddr()}
				cfgNewBrother := msg.NewBrothers(newBrother)
				oldBrothers := []string{}
				cfgNewInstance := c.GetCh().NewInstanceMessage(c.GetBindAddr(), c.GetName())
				fmt.Printf("\n newInstance : %#v\n", cfgNewInstance)
				for _, conn := range c.GetCh().GetConnsByName()[c.GetName()] {
					if conn.GetDir() == "in" && conn != c {
						c.GetCh().FSendConn("cfg")(conn, cfgNewInstance)
						oldBrothers = append(oldBrothers, conn.GetBindAddr())
						fmt.Printf("    Sending NewBrother %s to %s:%s\n", c.GetBindAddr(), c.GetName(), conn.GetBindAddr())
						c.GetCh().FSendConn("cfg")(conn, cfgNewBrother)
					}
				}
				cfgOldBrothers := msg.NewBrothers(oldBrothers)
				fmt.Printf("  Sending oldBrothers %#v to %s:%s \n", oldBrothers, c.GetName(), c.GetBindAddr())
				c.GetCh().FSendConn("cfg")(c, cfgOldBrothers)

				for connAddr := range c.GetCh().GetBrothers() {
					cfgConnectTo := c.GetCh().NewConnectToMessage(connAddr)
					c.GetCh().FSendConn("cfg")(c, cfgConnectTo)
				}
			}
		}
		if c.GetDir() == "out" {
			fmt.Printf("\n--\nhandhsake for %s to %s\n", c.GetCh().GetBindAddr(), c.GetBindAddr())
			newBrother := []string{c.GetBindAddr()}
			cfgNewBrother := msg.NewBrothers(newBrother)
			oldBrothers := []string{}
			for _, conn := range c.GetCh().GetConnsByName()[c.GetName()] {
				if conn.GetDir() == "out" && conn != c {
					oldBrothers = append(oldBrothers, conn.GetBindAddr())
					fmt.Printf("    Sending NewBrother %s to %s\n", c.GetBindAddr(), conn.GetBindAddr())
					c.GetCh().FSendConn("cfg")(conn, cfgNewBrother)
				}
			}
			fmt.Printf("  --> oldBrothers : %#v\n", oldBrothers)
			cfgOldBrothers := msg.NewBrothers(oldBrothers)
			fmt.Printf("  Sending oldBrothers to %s\n", c.GetName())
			c.GetCh().FSendConn("cfg")(c, cfgOldBrothers)
		}
	case "brothers":
		for _, addr := range cfg.Brothers {
			c.GetCh().GetBrothers()[addr] = true
		}
	case "newInstance":
		if c.GetDir() == "out" {
			cfgConnectTo := c.GetCh().NewConnectToMessage(cfg.Address)
			for _, conn := range c.GetCh().GetConnsByAddr() {
				if conn.GetDir() == "in" {
					c.GetCh().FSendConn("cfg")(conn, cfgConnectTo)
				}
			}
		}
	case "connectTo":
		if c.GetDir() == "out" {

			if c.GetCh().GetConnsByAddr()[cfg.Address] == nil {
				fmt.Printf("\n %s.connectTo : %s\n", c.GetCh().GetBindAddr(), cfg.Address)
				c.GetCh().Connect(cfg.Address)
			}
		}
	}
	return err
}

// SendConfigConn :
func SendConfigConn(c *ChaussetteConn, cfg interface{}) {
	fmt.Print("Sending config.\n")
	c.WriteString("cfg")
	c.WriteMessage(cfg)
}

// SendConfig :
func SendConfig(c *Chaussette, cfg interface{}) {
	fmt.Print("Not implemented !\n")
}

// WaitConfig :
func WaitConfig(c *Chaussette, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	fmt.Print("Not implemented !\n")
	return nil
}
