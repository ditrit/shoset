package shoset

import (
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
func (cbh *ConfigByeHandler) Handle(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol) // compute config from message
	ch := c.GetCh()

	switch cfg.GetCommandName() {
	case "bye":
		dir := c.GetDir()
		remoteAddress := cfg.GetAddress()
		if dir == "in" {
			cfgNewDelete := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "delete")
			ch.ConnsByName.IterateAll(
				func(address string, bro *ShosetConn) {
					if address != remoteAddress {
						if err := bro.SendMessage(*cfgNewDelete); err != nil {
							bro.ch.logger.Warn().Msg("couldn't send cfgnewdelete : " + err.Error())
						}
					}
				},
			)
			c.SetIsValid(false)
		}

	case "delete":
		ch.LnamesByProtocol.Set("bye", cfg.GetLogicalName())
		ch.LnamesByType.Set(cfg.GetShosetType(), cfg.GetLogicalName())
		ch.deleteConn(cfg.GetAddress(), cfg.GetLogicalName())
	}
	return nil
}

// SendConfigBye :
func (cbh *ConfigByeHandler) Send(c *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigByeHandler.Send not implemented")
	return
}

// WaitConfig :
func (cbh *ConfigByeHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigByeHandler.Wait not implemented")
	return nil
}
