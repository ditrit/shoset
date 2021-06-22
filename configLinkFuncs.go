package shoset

import (
	"github.com/ditrit/shoset/msg"
)

// GetConfigLink :
func GetConfigLink(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigLink
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigLink :
func HandleConfigLink(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigLink)
	ch := c.GetCh()
	dir := c.GetDir()
	//	err := c.ReadMessage(&cfg)
	switch cfg.GetCommandName() {
	case "handshake":
		if c.GetName() == "" {
			if dir == "in" {
				myConfigLink := NewHandshake(ch)
				c.SendMessage(*myConfigLink)
			}
			c.SetName(cfg.GetLogicalName())
			c.SetShosetType(cfg.GetShosetType())
		}
		if c.GetLocalAddress() == "" {
			c.SetLocalAddress(cfg.GetBindAddress())
		}
		if dir == "in" {
			ch.SetConn(c.GetLocalAddress(), c.GetShosetType(), c)
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
	}
	return nil
}

func sendCfgToBrothers(currentConn *ShosetConn) {
	ch := currentConn.GetCh()
	currentAddr := currentConn.GetRemoteAddress()
	currentName := currentConn.GetName()
	cfgNameBrothers := msg.NewNameBrothers([]string{currentAddr})
	oldNameBrothers := []string{}
	ch.ConnsByName.Iterate(currentName,
		func(key string, conn *ShosetConn) {
			if conn.GetDir() == "in" && conn.GetLocalAddress() != currentAddr {
				conn.SendMessage(cfgNameBrothers)
				oldNameBrothers = append(oldNameBrothers, conn.GetLocalAddress())
			}
		},
	)
	if len(oldNameBrothers) > 0 {
		currentConn.SendMessage(msg.NewNameBrothers(oldNameBrothers))
	}
}

//NewHandshake : Build a configLink Message
func NewHandshake(ch *Shoset) *msg.ConfigLink {
	return msg.NewHandshake(ch.bindAddress, ch.lName, ch.ShosetType)
}

//NewCfgOut : Build a configLink Message
func NewCfgOut(ch *Shoset) *msg.ConfigLink {
	var bros []string
	ch.ConnsByAddr.Iterate(
		func(key string, val *ShosetConn) {
			conn := val
			if conn.dir == "out" {
				bros = append(bros, conn.remoteAddr)
			}
		},
	)
	return msg.NewConns("out", bros)
}

//NewCfgIn : Build a configLink Message
func NewCfgIn(ch *Shoset) *msg.ConfigLink {
	var bros []string
	ch.Brothers.Iterate(
		func(key string, val bool) {
			bros = append(bros, key)
		},
	)
	return msg.NewConns("in", bros)
}
