package net

import (
	"shoset/msg"
)

// GetConfig :
func GetConfig(c *ShosetConn) (msg.Message, error) {
	var cfg msg.Config
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfig :
func HandleConfig(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.Config)
	ch := c.GetCh()
	dir := c.GetDir()
	//	err := c.ReadMessage(&cfg)
	switch cfg.GetCommandName() {
	case "handshake":
		if c.GetName() == "" {
			if dir == "in" {
				myConfig := NewHandshake(ch)
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
			cfgBros := NewCfgOut(ch)
			c.SendMessage(*cfgBros)
		}
	case "out":
		if dir == "in" {
			for _, bro := range cfg.Conns {
				ch.Brothers.Set(bro, true)
			}
			//???? pas d'envoi depuis ou vers conn ????
			cfgIn := NewCfgIn(ch)
			ch.ConnsByAddr.Iterate(
				func(key string, conn *ShosetConn) {
					c.SendMessage(cfgIn)
				},
			)
			sendCfgToBrothers(c)
		}
	case "in":
		if dir == "out" {
			for _, bro := range cfg.Conns {
				if ch.ConnsByAddr.Get(bro) == nil {
					ch.Link(bro)
				}
			}
			sendCfgToBrothers(c)
		}
	case "nameBrothers":
		for _, addr := range cfg.GetConns() {
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
		}
	case "connectTo":
		if dir == "out" {

			if ch.ConnsByAddr.Get(cfg.Address) == nil {
				ch.Link(cfg.Address)
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
	return nil
}

func sendCfgToBrothers(currentConn *ShosetConn) {
	ch := currentConn.GetCh()
	currentAddr := currentConn.GetBindAddr()
	currentName := currentConn.GetName()
	cfgNameBrothers := msg.NewNameBrothers([]string{currentAddr})
	oldNameBrothers := []string{}
	ch.ConnsByName.Iterate(currentName,
		func(key string, conn *ShosetConn) {
			if conn.GetDir() == "in" && conn.bindAddr != currentAddr {
				conn.SendMessage(cfgNameBrothers)
				oldNameBrothers = append(oldNameBrothers, conn.bindAddr)
			}
		},
	)
	if len(oldNameBrothers) > 0 {
		currentConn.SendMessage(msg.NewNameBrothers(oldNameBrothers))
	}
}

//NewHandshake : Build a config Message
func NewHandshake(ch *Shoset) *msg.Config {
	return msg.NewHandshake(ch.bindAddr, ch.lName, ch.ShosetType)
}

//NewCfgOut : Build a config Message
func NewCfgOut(ch *Shoset) *msg.Config {
	var bros []string
	ch.ConnsByAddr.Iterate(
		func(key string, val *ShosetConn) {
			conn := val
			if conn.dir == "out" {
				bros = append(bros, conn.addr)
			}
		},
	)
	return msg.NewConns("out", bros)
}

//NewCfgIn : Build a config Message
func NewCfgIn(ch *Shoset) *msg.Config {
	var bros []string
	ch.Brothers.Iterate(
		func(key string, val bool) {
			bros = append(bros, key)
		},
	)
	return msg.NewConns("in", bros)
}

// SendConfig :
func SendConfig(ch *Shoset, cfg msg.Message) {
}

// WaitConfig :
func WaitConfig(ch *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	return nil
}
