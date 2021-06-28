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
	remoteAddress := cfg.GetAddress()
	switch cfg.GetCommandName() {
	case "join":
		if dir == "in" { // a socket wants to join this one
			// fmt.Printf("########### a socket wants to join this one")
			if exists := c.ch.ConnsJoin.Get(remoteAddress); exists != nil {
				return nil
			}
			if ch.GetName() == cfg.GetName() && ch.GetShosetType() == cfg.GetShosetType() {
				// fmt.Printf("\n###########  same type")
				// fmt.Println("in : ", remoteAddress)
				c.SetRemoteAddress(remoteAddress)
				ch.ConnsJoin.Set(remoteAddress, c)
				ch.NameBrothers.Set(remoteAddress, true)
				// ch.Join(remoteAddress)
				configOk := msg.NewCfg(remoteAddress, ch.GetName(), ch.GetShosetType(), "ok")
				c.SendMessage(configOk)
			} else {
				// fmt.Println("Invalid connection for join - not the same type/name")
				c.SetIsValid(false) //////////////////////
				configNotOk := msg.NewCfg(remoteAddress, ch.GetName(), ch.GetShosetType(), "notok")
				c.SendMessage(configNotOk)
				// fmt.Println(c.GetIsValid(), " - after handleconfigjoin")
				return errors.New("error : Invalid connection for join - not the same type/name")
			}
		}

		thisOne := c.GetLocalAddress()
		cfgNewMember := msg.NewCfg(remoteAddress, ch.GetName(), ch.GetShosetType(), "member")
		ch.ConnsJoin.Iterate(
			func(key string, val *ShosetConn) {
				if key != remoteAddress && key != thisOne {
					val.SendMessage(cfgNewMember) //tell to the other member that there is a new member to join
				}
			},
		)

		// if dir == "out" {
		// }

	case "ok":
		// fmt.Println("ok : ", c.remoteAddr)
		ch.ConnsJoin.Set(c.GetRemoteAddress(), c) //////////////// need to find remote address because here we take the address of the tcp protocol which is random
		// fmt.Println("received ok", c.GetLocalAddress())
		ch.NameBrothers.Set(c.GetRemoteAddress(), true)

	case "notok":
		// fmt.Println("received notok", c.GetLocalAddress())
		c.SetIsValid(false)
		return errors.New("error : connection not ok")

	case "member":
		fmt.Println("message member recu")
		ch.Join(remoteAddress)
	}
	return nil
}
