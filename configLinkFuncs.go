package shoset

import (
	// "fmt"

	"github.com/ditrit/shoset/msg"
)

// GetConfigLink :
func GetConfigLink(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigLink :
func HandleConfigLink(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)
	remoteAddress := cfg.GetAddress()
	dir := c.GetDir()

	switch cfg.GetCommandName() {
	case "link":
		if dir == "in" { // a socket wants to link to this one
			if connsLink := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsLink != nil { //already linked
				if connsLink.Get(remoteAddress) != nil {
					return nil
				}
			}

			c.SetRemoteAddress(remoteAddress)
			c.SetRemoteLogicalName(cfg.GetLogicalName()) // avoid tcp port name
			c.SetRemoteShosetType(cfg.GetShosetType())
			c.ch.ConnsByName.Set(cfg.GetLogicalName(), remoteAddress, "link", cfg.GetShosetType(), c) // set conn in this socket
			// c.ch.LnamesByProtocol.Set("link", c.GetRemoteLogicalName())
			// c.ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

			localBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName())
			localBrothersArray := []string{}
			if localBrothers != nil {
				localBrothersArray = localBrothers.Keys("all")
			}

			remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
			remoteBrothersArray := []string{}
			if remoteBrothers != nil {
				remoteBrothersArray = remoteBrothers.Keys("all")
			}

			brothers := msg.NewCfgBrothers(localBrothersArray, remoteBrothersArray, c.ch.GetLogicalName(), "brothers", c.ch.GetShosetType())
			remoteBrothers.Iterate(
				func(address string, remoteBro *ShosetConn) {
					remoteBro.SendMessage(brothers) //send config to others
				},
			)
		}

	case "brothers":
		if dir == "out" { // this socket wants to link to another
			c.SetRemoteLogicalName(cfg.GetLogicalName())
			c.SetRemoteShosetType(cfg.GetShosetType())
			c.ch.ConnsByName.Set(cfg.GetLogicalName(), c.GetRemoteAddress(), "link", cfg.GetShosetType(), c) // set conns in the other socket
			// c.ch.LnamesByProtocol.Set("link", c.GetRemoteLogicalName())
			// c.ch.LnamesByType.Set(c.ch.GetShosetType(), c.GetRemoteLogicalName())

			localBrothers := cfg.GetYourBrothers()
			remoteBrothers := cfg.GetMyBrothers()

			for _, bro := range localBrothers {
				if bro != c.ch.GetBindAddress() {
					conn, err := NewShosetConn(c.ch, bro, "me") // create empty socket so that the two aga/Ca know each other
					if err == nil {
						conn.SetRemoteLogicalName(c.ch.GetLogicalName())
						conn.SetRemoteShosetType(c.ch.GetShosetType())
						c.ch.ConnsByName.Set(c.ch.GetLogicalName(), bro, "link", conn.GetRemoteShosetType(), conn) // musn't be linked !
					}

					newLocalBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName()).Keys("me")
					for _, lName := range c.ch.ConnsByName.Keys() {
						lNameConns := c.ch.ConnsByName.Get(lName)
						addresses := lNameConns.Keys("in")
						brothers := msg.NewCfgBrothers(newLocalBrothers, addresses, c.ch.GetLogicalName(), "brothers", c.ch.GetShosetType())
						lNameConns.Iterate(
							func(key string, val *ShosetConn) {
								val.SendMessage(brothers)
							})
					}
				}
			}

			for _, remoteBro := range remoteBrothers { // link to the brothers of the socket it's linked with
				remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
				if remoteBrothers != nil {
					if remoteBrothers.Get(remoteBro) == nil {
						c.ch.Protocol(remoteBro, "link")
					}
				}
			}
		}
	}
	return nil
}
