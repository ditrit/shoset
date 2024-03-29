package shoset

import (
	// "fmt"

	"github.com/ditrit/shoset/msg"
)

// GetConfigBye :
func GetConfigBye(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigBye :
func HandleConfigBye(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol) // compute config from message
	ch := c.GetCh()
	dir := c.GetDir()
	remoteAddress := cfg.GetAddress()

	switch cfg.GetCommandName() {
	case "bye":
		if dir == "in" {
			cfgNewDelete := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "delete")
			ch.ConnsByName.IterateAll(
				func(address string, bro *ShosetConn) {
					if address != remoteAddress {
						bro.SendMessage(cfgNewDelete)
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
