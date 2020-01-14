package net

import (
	"fmt"

	"../msg"
)

// HandleConfig :
func HandleConfig(c *ChaussetteConn) error {
	var cfg msg.Config
	ch := c.GetCh()
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
			c.GetCh().SetConn(c.GetBindAddr(), c)
		}
		if c.GetDir() == "out" {
			cfgBros := c.GetCh().NewCfgOut()
			c.SendMessage(*cfgBros)
		}
	case "out":
		if c.GetDir() == "in" {
			for _, bro := range cfg.Conns {
				ch.SetBrother(bro)
			}
			cfgIn := c.GetCh().NewCfgIn()
			for _, conn := range c.GetCh().GetConnsByAddr() {
				if conn.GetDir() == "in" {
					c.SendMessage(cfgIn)
				}
			}
		}
	case "in":
		if c.GetDir() == "out" {
			for _, bro := range cfg.Conns {
				fmt.Printf("| candidate bro = %s\n", bro)
				if c.GetCh().GetConnsByAddr()[bro] == nil {
					fmt.Printf(" - bro to add = %s\n", bro)
					c.GetCh().Connect(bro)
				}
			}
		}
	}
	return err
}

// SendConfig :
func SendConfig(c *Chaussette, cfg msg.Message) {
	fmt.Print("Not implemented !\n")
}

// WaitConfig :
func WaitConfig(c *Chaussette, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	fmt.Print("Not implemented !\n")
	return nil
}

/*
// HandleConfig :
func HandleConfigOld(c *ChaussetteConn) error {
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
						c.SendMessage(cfgNewInstance)
						oldBrothers = append(oldBrothers, conn.GetBindAddr())
						fmt.Printf("    Sending NewBrother %s to %s:%s\n", c.GetBindAddr(), c.GetName(), conn.GetBindAddr())
						c.SendMessage(cfgNewBrother)
					}
				}
				cfgOldBrothers := msg.NewBrothers(oldBrothers)
				fmt.Printf("  Sending oldBrothers %#v to %s:%s \n", oldBrothers, c.GetName(), c.GetBindAddr())
				c.SendMessage(cfgOldBrothers)

				for connAddr := range c.GetCh().GetBrothers() {
					cfgConnectTo := c.GetCh().NewConnectToMessage(connAddr)
					c.SendMessage(cfgConnectTo)
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
					c.SendMessage(cfgNewBrother)
				}
			}
			fmt.Printf("  --> oldBrothers : %#v\n", oldBrothers)
			cfgOldBrothers := msg.NewBrothers(oldBrothers)
			fmt.Printf("  Sending oldBrothers to %s\n", c.GetName())
			c.SendMessage(cfgOldBrothers)
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
					c.SendMessage(cfgConnectTo)
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
*/
