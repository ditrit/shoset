package shoset

import (
	"errors"
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
	// fmt.Printf("########### enter handleconfigjoin\n")
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
				fmt.Println("in : ", c.addr, cfg.GetBindAddress())
				ch.ConnsJoin.Set(cfg.GetBindAddress(), c)
				ch.NameBrothers.Set(cfg.GetBindAddress(), true)
				ch.Join(newMember)
				configOk := msg.NewCfg(newMember, ch.GetName(), ch.GetShosetType(), "ok")
				c.SendMessage(configOk)
			} else {
				fmt.Println("Invalid connection for join - not the same type/name")
				c.SetIsValid(false) //////////////////////
				fmt.Println(c.GetIsValid(), " - after handleconfigjoin")
				return errors.New("error : Invalid connection for join - not the same type/name")
			}
		}
		thisOne := c.bindAddr
		cfgNewMember := msg.NewCfg(newMember, ch.GetName(), ch.GetShosetType(), "member")
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
		fmt.Println("ok : ", c.addr)
		ch.ConnsJoin.Set(c.addr, c) //////////////// need to find remote address because here we take the address of the tcp protocol which is random
		ch.NameBrothers.Set(c.addr, true)

	case "member":
		newMember := cfg.GetBindAddress()
		ch.Join(newMember)
	}
	return nil
}
