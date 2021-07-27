package shoset

import (
	"errors"

	"github.com/ditrit/shoset/msg"
)

// GetConfigJoin :
func GetConfigJoin(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigJoin :
func HandleConfigJoin(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol) // compute config from message
	ch := c.GetCh()
	dir := c.GetDir()
	remoteAddress := cfg.GetAddress()

	switch cfg.GetCommandName() {
	case "join":
		if dir == "in" { // a socket wants to join this one
			if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already joined
				if connsJoin.Get(remoteAddress) != nil {
					return nil
				}
			}

			if ch.GetLogicalName() == cfg.GetLogicalName() && ch.GetShosetType() == cfg.GetShosetType() {
				c.SetRemoteAddress(remoteAddress)
				ch.ConnsByName.Set(ch.GetLogicalName(), remoteAddress, c) // set conn in this socket
				c.SetRemoteLogicalName(cfg.GetLogicalName())

				configOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "aknowledge_join")
				c.SendMessage(configOk)
			} else {
				c.SetIsValid(false)

				configNotOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "unaknowledge_join")
				c.SendMessage(configNotOk)
				return errors.New("error : Invalid connection for join - not the same type/name")
			}
		}

		cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
		ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
			func(address string, bro *ShosetConn) {
				if address != remoteAddress {
					bro.SendMessage(cfgNewMember) //tell to the other members that there is a new member to join
				}
			},
		)

	case "aknowledge_join":
		ch.ConnsByName.Set(ch.GetLogicalName(), c.GetRemoteAddress(), c) // set conns in the other socket
		c.SetRemoteLogicalName(cfg.GetLogicalName())

	case "unaknowledge_join":
		c.SetIsValid(false)
		return errors.New("error : connection not ok")

	case "member":
		if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already joined
			if connsJoin.Get(remoteAddress) == nil {
				ch.Protocol(remoteAddress, "join")

				cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
				ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
					func(address string, bro *ShosetConn) {
						if address != remoteAddress {
							bro.SendMessage(cfgNewMember) //tell to the other members that there is a new member to join
						}
					},
				)
			}
		}

	}
	return nil
}
