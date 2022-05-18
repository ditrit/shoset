package shoset

import (
	"errors"
	"sync"

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
func (clh *ConfigLinkHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)
	dir := c.GetDir()

	switch cfg.GetCommandName() {
	case PROTOCOL_LINK:
		// incoming link request
		// a socket wants to link to this one
		if dir != IN {
			return errors.New("wrong dir link")
		}

		remoteAddress := cfg.GetAddress()

		c.SetRemoteAddress(remoteAddress)
		c.SetRemoteLogicalName(cfg.GetLogicalName())
		c.SetRemoteShosetType(cfg.GetShosetType())
		c.ch.ConnsByName.Store(cfg.GetLogicalName(), remoteAddress, PROTOCOL_LINK, cfg.GetShosetType(), c.ch.config.GetFileName(), c)

		localBrothersArray := []string{}
		if localBrothers, _ := c.ch.ConnsByName.smap.Load(c.ch.GetLogicalName()); localBrothers != nil {
			localBrothersArray = Keys(localBrothers.(*sync.Map), ALL)
		}

		remoteBrothers, _ := c.ch.ConnsByName.smap.Load(cfg.GetLogicalName())
		remoteBrothersArray := []string{}
		if remoteBrothers != nil {
			remoteBrothersArray = Keys(remoteBrothers.(*sync.Map), ALL)
		}

		brothers := msg.NewCfgBrothers(localBrothersArray, remoteBrothersArray, c.ch.GetLogicalName(), BROTHERS, c.ch.GetShosetType())
		remoteBrothers.(*sync.Map).Range(func(key, value interface{}) bool {
			func(address string, remoteBro interface{}) {
				if err := remoteBro.(*ShosetConn).SendMessage(*brothers); err != nil {
					remoteBro.(*ShosetConn).logger.Warn().Msg("couldn't send brothers : " + err.Error())
				}
			}(key.(string), value)
			return true
		})

	case BROTHERS:
		// incoming brother information
		// new shoset in the network, must handle it
		if dir != OUT {
			return nil
		}

		c.SetRemoteLogicalName(cfg.GetLogicalName())
		c.SetRemoteShosetType(cfg.GetShosetType())
		c.ch.ConnsByName.Store(cfg.GetLogicalName(), c.GetRemoteAddress(), PROTOCOL_LINK, cfg.GetShosetType(), c.ch.config.GetFileName(), c)

		sendToBrothers(c, message)
	}
	return nil
}

// sendToBrothers : handle link brothers
func sendToBrothers(c *ShosetConn, m msg.Message) {
	cfg := m.(msg.ConfigProtocol)
	localBrothers := cfg.GetYourBrothers()
	remoteBrothers := cfg.GetMyBrothers()

	// handle local brothers (same type of shoset)
	// need to add them to our known shosets as "fake" connection
	for _, bro := range localBrothers {
		if bro == c.ch.GetBindAddress() {
			continue
		}

		conn, _ := NewShosetConn(c.ch, bro, ME)
		conn.SetRemoteLogicalName(c.ch.GetLogicalName())
		conn.SetRemoteShosetType(c.ch.GetShosetType())
		c.ch.ConnsByName.Store(c.ch.GetLogicalName(), bro, PROTOCOL_LINK, conn.GetRemoteShosetType(), c.ch.config.GetFileName(), conn)

		newLocalBrothers, _ := c.ch.ConnsByName.smap.Load(c.ch.GetLogicalName())
		newLocalBrothersList := Keys(newLocalBrothers.(*sync.Map), RELATIVE)
		for _, lName := range c.ch.ConnsByName.Keys(ALL) {
			lNameConns, _ := c.ch.ConnsByName.smap.Load(lName)
			addresses := Keys(lNameConns.(*sync.Map), IN)
			brothers := msg.NewCfgBrothers(newLocalBrothersList, addresses, c.ch.GetLogicalName(), BROTHERS, c.ch.GetShosetType())

			lNameConns.(*sync.Map).Range(func(key, value interface{}) bool {
				func(key string, val interface{}) {
					if err := val.(*ShosetConn).SendMessage(*brothers); err != nil {
						val.(*ShosetConn).logger.Warn().Msg("couldn't send newLocalBrothers : " + err.Error())
					}
				}(key.(string), value)
				return true
			})
		}
	}

	// handle remote brothers (not the same type of shoset)
	// need to link protocol on them
	for _, remoteBro := range remoteBrothers {
		remoteBrothers, _ := c.ch.ConnsByName.smap.Load(cfg.GetLogicalName())
		if remoteBrothers == nil {
			return
		}
		if exists, _ := remoteBrothers.(*sync.Map).Load(remoteBro); exists == nil {
			c.ch.Protocol(c.ch.GetBindAddress(), remoteBro, PROTOCOL_LINK)
		}
	}
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
