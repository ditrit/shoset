package shoset

import (
	"errors"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

type ConfigJoinHandler struct{}

// GetConfigJoin :
func (cjh *ConfigJoinHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigJoin :
func (cjh *ConfigJoinHandler) Handle(c *ShosetConn, message msg.Message) error {
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
				c.SetRemoteLogicalName(cfg.GetLogicalName())
				c.SetRemoteShosetType(cfg.GetShosetType())
				ch.ConnsByName.Set(ch.GetLogicalName(), remoteAddress, "join", ch.GetShosetType(), c.ch.GetFileName(), c) // set conn in this socket
				// ch.LnamesByProtocol.Set("join", c.GetRemoteLogicalName())
				// ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

				configOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "aknowledge_join")
				err := c.SendMessage(*configOk)
				if err != nil {
					c.ch.logger.Warn().Msg("couldn't send configOk : " + err.Error())
				}
			} else {
				c.SetIsValid(false)

				configNotOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "unaknowledge_join")
				err := c.SendMessage(*configNotOk)
				if err != nil {
					c.ch.logger.Warn().Msg("couldn't send configNotOk : " + err.Error())
				}
				return errors.New("error : Invalid connection for join - not the same type/name")
			}
		}

		cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
		ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
			func(address string, bro *ShosetConn) {
				if address != remoteAddress {
					//tell to the other members that there is a new member to join
					if err := bro.SendMessage(*cfgNewMember); err != nil {
						bro.ch.logger.Warn().Msg("couldn't send cfgnewMember : " + err.Error())
					}
				}
			},
		)

	case "aknowledge_join":
		c.SetRemoteLogicalName(cfg.GetLogicalName())
		c.SetRemoteShosetType(cfg.GetShosetType())
		ch.ConnsByName.Set(ch.GetLogicalName(), c.GetRemoteAddress(), "join", ch.GetShosetType(), c.ch.GetFileName(), c) // set conns in the other socket
		// c.ch.LnamesByProtocol.Set("join", c.GetRemoteLogicalName())
		// c.ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

	case "unaknowledge_join":
		c.SetIsValid(false)
		return errors.New("error : connection not ok")

	case "member":
		if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already joined
			if connsJoin.Get(remoteAddress) != nil {
				return nil
			}
			ch.Protocol(c.ch.GetBindAddress(), remoteAddress, "join")

			cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
			ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
				func(address string, bro *ShosetConn) {
					if address != remoteAddress {
						//tell to the other members that there is a new member to join
						if err := bro.SendMessage(*cfgNewMember); err != nil {
							bro.ch.logger.Warn().Msg("couldn't send cfgnewMember2 : " + err.Error())
						}
					}
				},
			)
		}

	}
	return nil
}

// SendConfigBye :
func (cjh *ConfigJoinHandler) Send(c *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Send not implemented")
}

// WaitConfig :
func (cjh *ConfigJoinHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Wait not implemented")
	return nil
}