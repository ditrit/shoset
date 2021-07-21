package shoset

import (
	"errors"
	"fmt"

	"github.com/ditrit/shoset/msg"
)

// GetConfigJoin :
func GetConfigJoin(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigJoin :
func HandleConfigJoin(c *ShosetConn, message msg.Message) error {
	fmt.Println("########### enter handleconfigjoin")
	cfg := message.(msg.ConfigProtocol) // compute config from message
	ch := c.GetCh()
	dir := c.GetDir()
	remoteAddress := cfg.GetAddress()
	switch cfg.GetCommandName() {
	case "join":
		if dir == "in" { // a socket wants to join this one
			if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already joined
				if connsJoin.Get(remoteAddress) != nil {
					return nil
				}
			}

			if ch.GetLogicalName() == cfg.GetLogicalName() && ch.GetShosetType() == cfg.GetShosetType() {
				// fmt.Printf("\n###########  same type")
				// fmt.Println("in : ", remoteAddress)
				c.SetRemoteAddress(remoteAddress)
				ch.ConnsByName.Set(ch.GetLogicalName(), remoteAddress, "join", c) // set conn in this socket
				c.SetRemoteLogicalName(cfg.GetLogicalName())
				// ch.Join(remoteAddress)
				configOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "aknowledge_join")
				c.SendMessage(configOk)
			} else {
				// fmt.Println("Invalid connection for join - not the same type/name")
				c.SetIsValid(false) //////////////////////
				configNotOk := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "unaknowledge_join")
				c.SendMessage(configNotOk)
				// fmt.Println(c.GetIsValid(), " - after handleconfigjoin")
				return errors.New("error : Invalid connection for join - not the same type/name")
			}
		}

		cfgNewMember := msg.NewCfg(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "member")
		ch.ConnsByName.Get(ch.GetLogicalName()).Iterate(
			func(address string, val *ShosetConn) {
				if address != remoteAddress && address != c.GetLocalAddress() {
					val.SendMessage(cfgNewMember) //tell to the other member that there is a new member to join
				}
			},
		)

	case "aknowledge_join":
		// fmt.Println("ok : ", c.remoteAddr)
		ch.ConnsByName.Set(ch.GetLogicalName(), c.GetRemoteAddress(), "join", c) // set conns in the other socket
		// fmt.Println("received ok", c.GetLocalAddress())
		c.SetRemoteLogicalName(cfg.GetLogicalName())

	case "unaknowledge_join":
		// fmt.Println("received notok", c.GetLocalAddress())
		c.SetIsValid(false)
		return errors.New("error : connection not ok")

	case "member":
		ch.Protocol(remoteAddress, "join")
	}
	return nil
}
