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
			if ch.NameBrothers.Len() > 0 {
				ch.NameBrothers.Iterate(
					func(bro string, val bool) {
						c.SendMessage(msg.NewConnectTo(bro))
					},
				)
			}
		}
		if dir == "out" {
			cfgBros := ch.NewCfgOut()
			c.SendMessage(*cfgBros)
		}
	case "out":
		if dir == "in" {
			for _, bro := range cfg.Conns {
				ch.Brothers.Set(bro, true)
			}
			//???? pas d'envoi depuis ou vers conn ????
			cfgIn := ch.NewCfgIn()
			ch.ConnsByAddr.Iterate(
				func(key string, conn *ShosetConn) {
					c.SendMessage(cfgIn)
				},
			)
			ch.SendCfgToBrothers(c)
		}
	case "in":
		if dir == "out" {
			for _, bro := range cfg.Conns {
				if ch.ConnsByAddr.Get(bro) == nil {
					ch.Connect(bro)
				}
			}
			ch.SendCfgToBrothers(c)
		}
	case "nameBrothers":
		for _, addr := range cfg.GetConns() {
			ch.ma.Lock()
			if !ch.NameBrothers.Get(addr) {
				ch.ConnsByAddr.Iterate(
					func(key string, val *ShosetConn) {
						conn := val
						if conn.GetDir() == "in" {
							conn.SendMessage(msg.NewConnectTo(addr))
						}
					},
				)
				ch.NameBrothers.Set(addr, true)
			}
			ch.ma.Unlock()
		}
	case "connectTo":
		if dir == "out" {

			if ch.ConnsByAddr.Get(cfg.Address) == nil {
				ch.Connect(cfg.Address)
			}
		}
	case "join":
		newMember := cfg.GetBindAddress() // recupere l'adresse distante

		if dir == "in" {
			ch.Join(newMember)
		}
		thisOne := c.bindAddr
		cfgNewMember := msg.NewCfgMember(newMember)
		ch.ConnsJoin.Iterate(
			func(key string, val *ShosetConn) {
				if key != newMember && key != thisOne {
					val.SendMessage(cfgNewMember)
				}
			},
		)

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
