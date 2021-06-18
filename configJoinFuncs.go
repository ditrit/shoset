package shoset

import (
	"fmt"

	"github.com/ditrit/shoset/msg"
)

// GetConfigJoin :
func GetConfigJoin(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigJoin
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigJoin :
func HandleConfigJoin(c *ShosetConn, message msg.Message) error {
	// fmt.Printf("########### enter handleconfigjion")
	cfg := message.(msg.ConfigJoin) // compute config from message
	ch := c.GetCh()
	dir := c.GetDir()
	switch cfg.GetCommandName() {
	case "join":
		newMember := cfg.GetBindAddress() // compute

		if dir == "in" { // a socket wants to join this one
			// fmt.Printf("########### a socket wants to join this one")
			if ch.GetName() == cfg.GetName() && ch.GetShosetType() == cfg.GetShosetType() {
				// fmt.Printf("\n###########  same type")
				ch.ConnsJoin.Set(c.addr, c)
				ch.NameBrothers.Set(c.addr, true)
				ch.Join(newMember)
				configOk := msg.NewCfgJoinOk()
				c.SendMessage(configOk)
			} else {
				fmt.Printf("error : Invalid connection for join - not the same type/name")
				c.SetIsValid(false)
				//
			}
		}
		thisOne := c.bindAddr
		cfgNewMember := msg.NewCfgMember(newMember, ch.GetName(), ch.GetShosetType())
		ch.ConnsJoin.Iterate(
			func(key string, val *ShosetConn) {
				if key != newMember && key != thisOne {
					val.SendMessage(cfgNewMember) //tell to the other member that there is a new member to join
				}
			},
		)

		// if dir == "out" {
		// }

	case "ok":
		ch.ConnsJoin.Set(c.addr, c)
		ch.NameBrothers.Set(c.addr, true)

	case "member":
		newMember := cfg.GetBindAddress()
		ch.Join(newMember)
	}
	return nil
}
