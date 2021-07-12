package shoset

import (
	"fmt"

	"github.com/ditrit/shoset/msg"
)

// GetConfigLink :
func GetConfigLink(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigLink
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigLink :
func HandleConfigLink(c *ShosetConn, message msg.Message) error {
	fmt.Println("enter handleconfiglink !!!")
	cfg := message.(msg.ConfigLink)
	remoteAddress := cfg.GetAddress()
	dir := c.GetDir()
	switch cfg.GetCommandName() {
	case "link":
		if dir == "in" { // a socket wants to link to this one
			if connsJoin := c.ch.ConnsByName.Get(c.ch.GetLogicalName()); connsJoin != nil { //already linked
				if connsJoin.Get(remoteAddress) != nil {
					return nil
				}
			}

			c.SetRemoteAddress(remoteAddress)
			c.ch.ConnsByName.Set(c.ch.GetLogicalName(), remoteAddress, c) // set conn in this socket

			fmt.Println(c.ch.ConnsByName)
			localBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName())
			localBrothersArray := []string{}
			if localBrothers != nil {
				localBrothersArray = localBrothers.Keys()
			}

			remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
			remoteBrothersArray := []string{}
			if remoteBrothers != nil {
				remoteBrothersArray = remoteBrothers.Keys()
			}

			fmt.Println(localBrothersArray, remoteBrothersArray)

			aknowledge_brothers := msg.NewCfgBrothers(localBrothersArray, remoteBrothersArray)
			c.SendMessage(aknowledge_brothers)
		}

	case "brothers":
		if dir == "out" { // this socket wants to link to another
			c.ch.ConnsByName.Set(c.ch.GetLogicalName(), c.GetRemoteAddress(), c) // set conns in the other socket

			localBrothers := cfg.GetYourBrothers()
			myKnownLocalBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName())
			remoteBrothers := cfg.GetMyBrothers()
			// myKnownRemoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())

			for _, bro := range localBrothers {
				if myKnownLocalBrothers.Get(bro) == nil {
					conn, err := NewShosetConn(c.ch, bro, "me")
					if err != nil {
						fmt.Println("!!!!!!!!!!! new bro")
						c.ch.ConnsByName.Set(c.ch.GetLogicalName(), bro, conn)
					}	
					// for _, bro := range c.ch.ConnsByName.Keys() {

					// }
				}
			}

			for _, remoteBro := range remoteBrothers {
				fmt.Println("!!!!!!!!!!! new link")
				c.ch.Link(remoteBro)
			}
		}
	}
	return nil
}
