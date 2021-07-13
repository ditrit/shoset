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
	// fmt.Println("enter handleconfiglink !!!")
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

			c.SetRemoteAddress(remoteAddress)                            // avoid tcp port name
			c.ch.ConnsByName.Set(cfg.GetLogicalName(), remoteAddress, c) // set conn in this socket
			c.SetName(cfg.GetLogicalName())

			// fmt.Println(c.ch)
			// fmt.Println(c.ch.ConnsByName)

			localBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName())
			// fmt.Println("local brothers : ", localBrothers)
			localBrothersArray := []string{}
			if localBrothers != nil {
				localBrothersArray = localBrothers.Keys()
			}

			remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
			// fmt.Println("remote brothers : ", remoteBrothers)
			remoteBrothersArray := []string{}
			if remoteBrothers != nil {
				remoteBrothersArray = remoteBrothers.Keys()
			}

			// fmt.Println("brothers arrays : ", localBrothersArray, remoteBrothersArray)

			aknowledge_brothers := msg.NewCfgBrothers(localBrothersArray, remoteBrothersArray, c.ch.GetLogicalName())
			c.SendMessage(aknowledge_brothers)
		}

	case "brothers":
		// fmt.Println(c.ch.GetBindAddress(), " enter case brothers")
		if dir == "out" { // this socket wants to link to another
			// fmt.Println("config name : ", cfg.GetLogicalName())
			c.ch.ConnsByName.Set(cfg.GetLogicalName(), c.GetRemoteAddress(), c) // set conns in the other socket
			// c.ch.ConnsByName.Set(c.ch.GetLogicalName(), c.GetRemoteAddress(), c) // set conns in the other socket
			c.SetName(cfg.GetLogicalName())

			// fmt.Println(c.ch)
			// fmt.Println(c.ch.ConnsByName)

			localBrothers := cfg.GetYourBrothers()
			// fmt.Println("local brothers 2 : ", localBrothers)
			myKnownLocalBrothers := c.ch.ConnsByName.Get(c.ch.GetLogicalName())
			// fmt.Println("I'm ", c.ch.GetBindAddress(), " and here are my known local brothers : ", myKnownLocalBrothers)
			remoteBrothers := cfg.GetMyBrothers()
			// fmt.Println("remote brothers 2 : ", remoteBrothers)
			// myKnownRemoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
			if myKnownLocalBrothers != nil {
				for _, bro := range localBrothers { // à tester en rajoutant un aga ////////////////// ne fonctionne pas encore
					if myKnownLocalBrothers.Get(bro) == nil && bro != c.ch.GetBindAddress() {
						conn, err := NewShosetConn(c.ch, bro, "me") // create empty socket so that the two aga know each other
						if err != nil {
							fmt.Println("!!!!!!!!!!! new bro")
							c.ch.ConnsByName.Set(c.ch.GetLogicalName(), bro, conn) // put them into ConnsByName
						}
						// for _, bro := range c.ch.ConnsByName.Keys() {

						// }
					}
				}
			}

			for _, remoteBro := range remoteBrothers { // link to the brothers of the socket it's linked with
				remoteBrothers := c.ch.ConnsByName.Get(cfg.GetLogicalName())
				if remoteBrothers != nil {
					if remoteBrothers.Get(remoteBro) == nil {
						fmt.Println("!!!!!!!!!!! new link")
						c.ch.Link(remoteBro)
					}
				}
			}
		}
	}
	return nil
}
