package shoset

import (
	"fmt"

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
	fmt.Println("enter byeeeee")
	cfg := message.(msg.ConfigProtocol) // compute config from message
	ch := c.GetCh()
	dir := c.GetDir()
	remoteAddress := cfg.GetAddress()

	switch cfg.GetCommandName() {
	case "bye":
		fmt.Println("bye")
		if dir == "in" {
			cfgNewDelete := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "delete")
			fmt.Println("cfg ok")
			ch.ConnsByName.IterateAll(
				func(address string, bro *ShosetConn) {
					fmt.Println("+")
					if address != remoteAddress {
						fmt.Println("sendmessage")
						bro.SendMessage(cfgNewDelete)
					}
				},
			)

			// setIsValid(false) ??

		}

	case "delete":
		fmt.Println("delete")
		if dir == "out" {
			ch.deleteConn(cfg.GetAddress(), cfg.GetLogicalName())
		}
	}
	return nil
}
