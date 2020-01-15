package net

import (
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
	}
	return err
}

// SendConfig :
func SendConfig(c *Chaussette, cfg msg.Message) {
}

// WaitConfig :
func WaitConfig(c *Chaussette, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	return nil
}
