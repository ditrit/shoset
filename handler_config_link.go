package shoset

import (
	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

type ConfigLinkHandler struct{}

// GetConfigLink :
func (clh *ConfigLinkHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigLink :
func (clh *ConfigLinkHandler) Handle(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)
	dir := c.GetDir()

	switch cfg.GetCommandName() {
	case "link":
		remoteAddress := cfg.GetAddress()
		if dir == "in" { // a socket wants to link to this one
			if connsLink := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsLink != nil { //already linked
				if connsLink.Get(remoteAddress) != nil {
					return nil
				}
			}

			c.SetRemoteAddress(remoteAddress)
			c.SetRemoteLogicalName(cfg.GetLogicalName()) // avoid tcp port name
			c.SetRemoteShosetType(cfg.GetShosetType())
			c.ch.ConnsByName.Set(cfg.GetLogicalName(), remoteAddress, "link", cfg.GetShosetType(), c.ch.GetFileName(), c) // set conn in this socket
			// c.ch.LnamesByProtocol.Set("link", c.GetRemoteLogicalName())
			// c.ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

			localBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName())
			localBrothersArray := []string{}
			if localBrothers != nil {
				localBrothersArray = localBrothers.Keys("all")
			}

			remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
			remoteBrothersArray := []string{}
			if remoteBrothers != nil {
				remoteBrothersArray = remoteBrothers.Keys("all")
			}

			brothers := msg.NewCfgBrothers(localBrothersArray, remoteBrothersArray, c.ch.GetLogicalName(), "brothers", c.ch.GetShosetType())
			remoteBrothers.Iterate(
				func(address string, remoteBro *ShosetConn) {
					//send config to others
					if err := remoteBro.SendMessage(*brothers); err != nil {
						remoteBro.ch.logger.Warn().Msg("couldn't send brothers : " + err.Error())
					}
				},
			)
		}

	case "brothers":
		if dir == "out" { // this socket wants to link to another
			c.SetRemoteLogicalName(cfg.GetLogicalName())
			c.SetRemoteShosetType(cfg.GetShosetType())
			c.ch.ConnsByName.Set(cfg.GetLogicalName(), c.GetRemoteAddress(), "link", cfg.GetShosetType(), c.ch.GetFileName(), c) // set conns in the other socket
			// c.ch.LnamesByProtocol.Set("link", c.GetRemoteLogicalName())
			// c.ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

			localBrothers := cfg.GetYourBrothers()
			remoteBrothers := cfg.GetMyBrothers()

			for _, bro := range localBrothers {
				if bro != c.ch.GetBindAddress() {
					conn, err := NewShosetConn(c.ch, bro, "me") // create empty socket so that the two aga/Ca know each other
					if err == nil {
						conn.SetRemoteLogicalName(c.ch.GetLogicalName())
						conn.SetRemoteShosetType(c.ch.GetShosetType())
						c.ch.ConnsByName.Set(c.ch.GetLogicalName(), bro, "link", conn.GetRemoteShosetType(), c.ch.GetFileName(), conn) // musn't be linked !
					}

					newLocalBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName()).Keys("me")
					for _, lName := range c.ch.ConnsByName.Keys() {
						lNameConns := c.ch.ConnsByName.Get(lName)
						addresses := lNameConns.Keys("in")
						brothers := msg.NewCfgBrothers(newLocalBrothers, addresses, c.ch.GetLogicalName(), "brothers", c.ch.GetShosetType())
						lNameConns.Iterate(
							func(key string, val *ShosetConn) {
								// val.rb = msg.NewReader(c.socket)
								// val.wb = msg.NewWriter(c.socket)
								if err := val.SendMessage(*brothers); err != nil {
									val.ch.logger.Warn().Msg("couldn't send newLocalBrothers : " + err.Error())
								}
							})
					}
				}
			}

			for _, remoteBro := range remoteBrothers { // link to the brothers of the socket it's linked with
				remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
				if remoteBrothers != nil {
					if remoteBrothers.Get(remoteBro) == nil {
						c.ch.Protocol(c.ch.GetBindAddress(), remoteBro, "link")
					}
				}
			}
		}
	}
	return nil
}

// SendConfigLink :
func (clh *ConfigLinkHandler) Send(c *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigLinkHandler.Send not implemented")
}

// WaitConfigLink :
func (clh *ConfigLinkHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigLinkHandler.Wait not implemented")
	return nil
}
