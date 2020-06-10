package shoset

import (
	"github.com/mathieucaroff/shoset/msg"
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

		if dir == "in" {
			ch.Join(newMember)
		}
		thisOne := c.bindAddr
		cfgNewMember := msg.NewCfgMember(newMember)
		ch.ConnsJoin.Iterate(
			func(key string, val *ShosetConn) {
				if key != newMember && key != thisOne {
					val.SendMessage(cfgNewMember)
				}
			},
		)

		if dir == "out" {
		}

	case "member":
		newMember := cfg.GetBindAddress()
		ch.Join(newMember)
	}
	return nil
}
