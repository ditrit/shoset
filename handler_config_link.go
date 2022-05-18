package shoset

import (
	"sync"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// ConfigLinkHandler implements MessageHandlers interface.
type ConfigLinkHandler struct{}

// Get returns the message for a given ShosetConn.
func (clh *ConfigLinkHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.GetReader().ReadMessage(&cfg)
	return cfg, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (clh *ConfigLinkHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)

	switch cfg.GetCommandName() {
	case PROTOCOL_LINK:
		// incoming link request, a socket wants to link to this one.
		// save info and retrieve brothers to inform network.
		c.SetRemoteAddress(cfg.GetAddress())
		c.Store(PROTOCOL_LINK, cfg.GetLogicalName(), cfg.GetAddress(), cfg.GetShosetType())

		localBrothersArray := []string{}
		if localBrothers, _ := c.GetShoset().ConnsByLname.smap.Load(c.GetShoset().GetLogicalName()); localBrothers != nil {
			localBrothersArray = Keys(localBrothers.(*sync.Map), ALL)
		}

		remoteBrothers, _ := c.GetShoset().ConnsByLname.smap.Load(c.GetRemoteLogicalName())
		remoteBrothersArray := []string{}
		if remoteBrothers != nil {
			remoteBrothersArray = Keys(remoteBrothers.(*sync.Map), ALL)
		}

		cfgBrothers := msg.NewCfgBrothers(localBrothersArray, remoteBrothersArray, BROTHERS, c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType())
		remoteBrothers.(*sync.Map).Range(func(_, value interface{}) bool {
			func(_, remoteBro interface{}) {
				if err := remoteBro.(*ShosetConn).SendMessage(*cfgBrothers); err != nil {
					remoteBro.(*ShosetConn).Logger.Warn().Msg("couldn't send brothers : " + err.Error())
				}
			}(nil, value)
			return true
		})

	case BROTHERS:
		// incoming brother information, new shoset in the network.
		// save info and call sendToBrothers to handle message.
		c.Store(PROTOCOL_LINK, cfg.GetLogicalName(), c.GetRemoteAddress(), cfg.GetShosetType())

		sendToBrothers(c, message)
	}
	return nil
}

// sendToBrothers : handle link brothers.
// retrieve info concerning local and remote brothers and handle them.
func sendToBrothers(c *ShosetConn, m msg.Message) {
	cfg := m.(msg.ConfigProtocol)
	localBrothers := cfg.GetYourBrothers()
	remoteBrothers := cfg.GetMyBrothers()

	// handle local brothers (same type of shoset).
	// need to add them to our known shosets as "fake" connection but do not protocol on it.
	for _, bro := range localBrothers {
		if bro == c.GetShoset().GetBindAddress() {
			continue
		}

		fake_conn, _ := NewShosetConn(c.GetShoset(), bro, ME)
		fake_conn.SetRemoteLogicalName(c.GetLocalLogicalName())
		fake_conn.SetRemoteShosetType(c.GetLocalShosetType())

		mapSync := new(sync.Map)
		mapSync.Store(c.GetLocalLogicalName(), true)
		c.GetShoset().LnamesByProtocol.smap.Store(PROTOCOL_LINK, mapSync)
		c.GetShoset().LnamesByType.smap.Store(c.GetRemoteShosetType(), mapSync)
		c.GetShoset().ConnsByLname.Store(c.GetShoset().GetLogicalName(), bro, PROTOCOL_LINK, fake_conn)

	}

	// handle remote brothers (not the same type of shoset).
	// need to link protocol on them if not already in the map of known conn.
	mapConns, _ := c.GetShoset().ConnsByLname.smap.Load(cfg.GetLogicalName())
	if mapConns == nil {
		return
	}
	for _, remoteBrother := range remoteBrothers {
		if exists, _ := mapConns.(*sync.Map).Load(remoteBrother); exists == nil {
			c.GetShoset().Protocol(c.GetShoset().GetBindAddress(), remoteBrother, PROTOCOL_LINK)
		}
	}
}

// Send sends the message through the given Shoset network.
func (clh *ConfigLinkHandler) Send(s *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigLinkHandler.Send not implemented")
}

// Wait returns the message received for a given Shoset.
func (clh *ConfigLinkHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigLinkHandler.Wait not implemented")
	return nil
}
