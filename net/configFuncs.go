package net

import (
	"shoset/msg"
)

// HandleConfig :
func HandleConfig(c *ShosetConn) error {
	var cfg msg.Config
	ch := c.GetCh()
	dir := c.GetDir()
	err := c.ReadMessage(&cfg)
	switch cfg.GetCommandName() {
	case "handshake":
		if c.GetName() == "" {
			if dir == "in" {
				myConfig := ch.NewHandshake()
				c.SendMessage(*myConfig)
			}
			c.SetName(cfg.GetLogicalName())
			c.SetShosetType(cfg.GetShosetType())
		}
		if c.GetBindAddr() == "" {
			c.SetBindAddr(cfg.GetBindAddress())
		}
		if dir == "in" {
			ch.SetConn(c.GetBindAddr(), c.GetShosetType(), c)
			ch.SendConnectToNameBrothers(c)
		}
		if dir == "out" {
			cfgBros := ch.NewCfgOut()
			c.SendMessage(*cfgBros)
		}
	case "out":
		if dir == "in" {
			for _, bro := range cfg.Conns {
				ch.SetBrother(bro)
			}
			ch.SendCfgInToInConnsByAddr(c)
			ch.SendCfgToBrothers(c)
		}
	case "in":
		if dir == "out" {
			for _, bro := range cfg.Conns {
				if ch.GetConnByAddr(bro) == nil {
					ch.Connect(bro)
				}
			}
			ch.SendCfgToBrothers(c)
		}
	case "nameBrothers":
		for _, addr := range cfg.GetConns() {
			ch.SendConnectToInConnsByAddr(addr)
		}
	case "connectTo":
		if dir == "out" {

			if ch.GetConnByAddr(cfg.Address) == nil {
				ch.Connect(cfg.Address)
			}
		}
	case "join":
		newMember := cfg.GetBindAddress() // recupere l'adresse distante

		if dir == "in" {
			ch.Join(newMember)
		}
		ch.SendNewMemberToConnsJoin(newMember)

		if dir == "out" {
		}

	case "member":
		newMember := cfg.GetBindAddress()
		ch.Join(newMember)
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
