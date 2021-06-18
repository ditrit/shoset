package shoset

import (
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
	cfg := message.(msg.ConfigJoin)
	ch := c.GetCh()
	dir := c.GetDir()
	switch cfg.GetCommandName() {
	case "join":
		newMember := cfg.GetBindAddress() // recupere l'adresse distante

		if dir == "in" { // je suis une chaussette et je recois une demande join - je veux me joiner a toi
			if ch.GetName() == cfg.GetName() && ch.GetShosetType() == cfg.GetShosetType() {
				ch.ConnsJoin.Set(c.addr, c)
				ch.NameBrothers.Set(c.addr, true)
				ch.Join(newMember)
				configOk := msg.NewCfgJoinOk()
				c.SendMessage(configOk)
			} else {
				c.SetIsValid(false)
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

		if dir == "out" {
		}

	case "ok":
		ch.ConnsJoin.Set(c.addr, c)
		ch.NameBrothers.Set(c.addr, true)

	case "member":
		newMember := cfg.GetBindAddress()
		ch.Join(newMember)
	}
	return nil
}
