package net

import (
	"shoset/msg"
)

// HandleConfig :
func HandleConfig(c *ShosetConn) error {
	var cfg msg.Config
	ch := c.GetCh()
	err := c.ReadMessage(&cfg)
	switch cfg.GetCommandName() {
	case "handshake":
		if c.GetName() == "" {
			if c.GetDir() == "in" {
				myConfig := c.GetCh().NewHandshake()
				c.SendMessage(*myConfig)
			}
			c.SetName(cfg.GetLogicalName())
		}
		if c.GetBindAddr() == "" {
			c.SetBindAddr(cfg.GetBindAddress())
		}
		if c.GetDir() == "in" {
			c.GetCh().SetConn(c.GetBindAddr(), c)
			if len(c.GetCh().GetNameBrothers()) > 0 {
				for bro := range c.GetCh().GetNameBrothers() {
					c.SendMessage(msg.NewConnectTo(bro))
				}
			}
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
			cfgNameBrothers := msg.NewNameBrothers([]string{c.GetBindAddr()})
			oldNameBrothers := []string{}
			for _, conn := range c.GetCh().GetConnsByName()[c.GetName()] {
				if conn.GetDir() == "in" && conn.bindAddr != c.GetBindAddr() {
					conn.SendMessage(cfgNameBrothers)
					oldNameBrothers = append(oldNameBrothers, conn.bindAddr)
				}
			}
			if len(oldNameBrothers) > 0 {
				c.SendMessage(msg.NewNameBrothers(oldNameBrothers))
			}
		}
	case "in":
		if c.GetDir() == "out" {
			for _, bro := range cfg.Conns {
				if c.GetCh().GetConnsByAddr()[bro] == nil {
					c.GetCh().Connect(bro)
				}
			}
			cfgNameBrothers := msg.NewNameBrothers([]string{c.GetBindAddr()})
			oldNameBrothers := []string{}
			for _, conn := range c.GetCh().GetConnsByName()[c.GetName()] {
				if conn.GetDir() == "out" && conn.bindAddr != c.GetBindAddr() {
					conn.SendMessage(cfgNameBrothers)
					oldNameBrothers = append(oldNameBrothers, conn.bindAddr)
				}
			}
			if len(oldNameBrothers) > 0 {
				c.SendMessage(msg.NewNameBrothers(oldNameBrothers))
			}
		}
	case "nameBrothers":
		for _, addr := range cfg.GetConns() {
			if !c.GetCh().InNameBrothers(addr) {
				for _, conn := range c.GetCh().GetConnsByAddr() {
					if conn.GetDir() == "in" {
						conn.SendMessage(msg.NewConnectTo(addr))
					}
				}
				c.GetCh().SetNameBrother(addr)
			}
		}
	case "connectTo":
		if c.GetDir() == "out" {

			if c.GetCh().GetConnsByAddr()[cfg.Address] == nil {
				c.GetCh().Connect(cfg.Address)
			}
		}
	case "join":
		newMember := cfg.GetBindAddress()  // recupere l'adresse distante
		thisOne := c.GetCh().GetBindAddr() // adresse locale
		if c.GetDir() == "in" {
			if !c.GetCh().InConnsJoin(newMember) && newMember != thisOne {
				c.GetCh().Join(newMember)
			}
		}
		cfgNewMember := msg.NewCfgMember(newMember)
		for connAddr, conn := range c.GetCh().GetConnsJoin() {
			if connAddr != newMember && connAddr != thisOne {
				conn.SendMessage(cfgNewMember)
			}
		}
		if c.GetDir() == "out" {
		}

	case "member":
		newMember := cfg.GetBindAddress()
		thisOne := c.GetCh().GetBindAddr()
		if !c.GetCh().InConnsJoin(newMember) && newMember != thisOne {
			c.GetCh().Join(newMember)
		}
	}
	return err
}

// SendConfig :
func SendConfig(c *Shoset, cfg msg.Message) {
}

// WaitConfig :
func WaitConfig(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	return nil
}
