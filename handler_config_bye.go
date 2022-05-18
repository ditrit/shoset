package shoset

import (
	"sync"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

type ConfigByeHandler struct{}

// GetConfigBye :
func (cbh *ConfigByeHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigBye :
func (cbh *ConfigByeHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)
	ch := c.GetCh()

	switch cfg.GetCommandName() {
	case PROTOCOL_EXIT:
		// incoming bye request
		// send delete signal to all connected shosets from our list of known shosets
		dir := c.GetDir()
		remoteAddress := cfg.GetAddress()
		if dir == IN {
			cfgNewDelete := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), DELETE)
			ch.ConnsByName.Iterate(
				func(address string, bro interface{}) {
					if address != remoteAddress {
						if err := bro.(*ShosetConn).SendMessage(*cfgNewDelete); err != nil {
							bro.(*ShosetConn).logger.Warn().Msg("couldn't send cfgnewdelete : " + err.Error())
						}
					}
				},
			)
		}

	case DELETE:
		// incoming delete signal
		// forget the concerned shoset from our list of known shosets
		mapSync := new(sync.Map)
		mapSync.Store(cfg.GetLogicalName(), true)
		ch.LnamesByProtocol.smap.Store(PROTOCOL_EXIT, mapSync)
		ch.LnamesByType.smap.Store(cfg.GetShosetType(), mapSync)
		ch.deleteConn(cfg.GetAddress(), cfg.GetLogicalName())
	}
	return nil
}

// SendConfigBye :
func (cbh *ConfigByeHandler) Send(c *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigByeHandler.Send not implemented")
}

// WaitConfig :
func (cbh *ConfigByeHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigByeHandler.Wait not implemented")
	return nil
}
