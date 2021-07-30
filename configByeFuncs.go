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

			c.SetRemoteAddress(remoteAddress)
			c.SetRemoteLogicalName(cfg.GetLogicalName())
			ch.ConnsByName.Set(ch.GetLogicalName(), remoteAddress, "bye", c) // set conn in this socket

			configOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "aknowledge_Bye")
			c.SendMessage(configOk)
		}

		cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
		ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
			func(address string, bro *ShosetConn) {
				if address != remoteAddress {
					bro.SendMessage(cfgNewMember) //tell to the other members that there is a new member to Bye
				}
			},
		)

	case "aknowledge_Bye":
		c.SetRemoteLogicalName(cfg.GetLogicalName())
		ch.ConnsByName.Set(ch.GetLogicalName(), c.GetRemoteAddress(), "bye", c) // set conns in the other socket

	case "member":
		if connsBye := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsBye != nil { //already Byeed
			if connsBye.Get(remoteAddress) == nil {
				ch.Protocol(remoteAddress, "bye")

				cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
				ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
					func(address string, bro *ShosetConn) {
						if address != remoteAddress {
							bro.SendMessage(cfgNewMember) //tell to the other members that there is a new member to Bye
						}
					},
				)
			}
		}

	}
	return nil
}
