package shoset

import (
	"errors"
	// "fmt"

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
		// fmt.Println(c.ch.GetBindAddress(), "enters join for ", remoteAddress)
		if dir == "in" { // a socket wants to join this one
			if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already joined
				if connsJoin.Get(remoteAddress) != nil {
					return nil
				}
			}

			if ch.GetLogicalName() == cfg.GetLogicalName() && ch.GetShosetType() == cfg.GetShosetType() {
				c.SetRemoteAddress(remoteAddress)
				c.SetRemoteLogicalName(cfg.GetLogicalName())
				c.SetRemoteShosetType(cfg.GetShosetType())
				ch.ConnsByName.Set(ch.GetLogicalName(), remoteAddress, "join", ch.GetShosetType(), c) // set conn in this socket
				// ch.LnamesByProtocol.Set("join", c.GetRemoteLogicalName())
				// ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

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
		c.SetRemoteLogicalName(cfg.GetLogicalName())
		c.SetRemoteShosetType(cfg.GetShosetType())
		ch.ConnsByName.Set(ch.GetLogicalName(), c.GetRemoteAddress(), "join", ch.GetShosetType(), c) // set conns in the other socket
		// c.ch.LnamesByProtocol.Set("join", c.GetRemoteLogicalName())
		// c.ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

	case "unaknowledge_join":
		c.SetIsValid(false)
		return errors.New("error : connection not ok")

	case "member":
		if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already joined
			if connsJoin.Get(remoteAddress) == nil {
				ch.Protocol(c.ch.GetBindAddress(), remoteAddress, "join")

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
