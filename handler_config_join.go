package shoset

import (
	"sync"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// ConfigJoinHandler implements MessageHandlers interface.
type ConfigJoinHandler struct{}

// Get returns the message for a given ShosetConn.
func (cjh *ConfigJoinHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.GetReader().ReadMessage(&cfg)
	return cfg, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (cjh *ConfigJoinHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)

	switch cfg.GetCommandName() {
	case PROTOCOL_JOIN:
		// incoming join request, a socket wants to join to this one.
		// save info and retrieve brothers to inform network.

		//fmt.Println("PROTOCOL_JOIN")

		// Delete any existing conn
		c.GetShoset().deleteConn(cfg.GetLogicalName(), cfg.GetAddress())

		c.SetRemoteAddress(cfg.GetAddress())
		c.Store(PROTOCOL_JOIN, c.GetShoset().GetLogicalName(), cfg.GetAddress(), c.GetShoset().GetShosetType())

		configOk := msg.NewConfigProtocol(cfg.GetAddress(), c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType(), ACKNOWLEDGE_JOIN)
		if err := c.GetWriter().SendMessage(*configOk); err != nil {
			c.Logger.Warn().Msg("couldn't send configOk : " + err.Error())
		}

		cfgNewMember := msg.NewConfigProtocol(cfg.GetAddress(), c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType(), MEMBER)
		mapConns, _ := c.GetShoset().ConnsByLname.Load(c.GetShoset().GetLogicalName())
		mapConns.(*sync.Map).Range(func(key, value interface{}) bool {
			func(address string, bro interface{}) {
				if address == cfg.GetAddress() {
					return
				}
				if err := bro.(*ShosetConn).GetWriter().SendMessage(*cfgNewMember); err != nil {
					bro.(*ShosetConn).Logger.Warn().Msg("couldn't send cfgNewMember : " + err.Error())
				}
			}(key.(string), value)
			return true
		})

		//c.SetIsValid(true) // Send statusChange Event change status
		//c.GetShoset().waitGroupProtocol.Done()

	case ACKNOWLEDGE_JOIN:
		// incoming acknowledge_join, join request validated.
		// save info.

		//fmt.Println("ACKNOWLEDGE_JOIN")

		// Delete any existing conn
		c.GetShoset().deleteConn(cfg.GetLogicalName(), cfg.GetAddress())

		c.Store(PROTOCOL_JOIN, c.GetShoset().GetLogicalName(), c.GetRemoteAddress(), c.GetShoset().GetShosetType())

		//c.SetIsValid(true) // Send statusChange Event change status
		//c.GetShoset().waitGroupProtocol.Done()
		c.GetShoset().LaunchedProtocol.DeleteFromConcurentSlice(c.GetRemoteAddress())

	case MEMBER:
		// incoming member information.
		// need to link protocol on it if not already in the map of known conn.

		//fmt.Println("MEMBER")

		mapConns, _ := c.GetShoset().ConnsByLname.Load(c.GetShoset().GetLogicalName())
		if mapConns == nil {
			return nil
		}
		if exists, _ := mapConns.(*sync.Map).Load(cfg.GetAddress()); exists == nil {
			c.GetShoset().Protocol(c.GetShoset().GetBindAddress(), cfg.GetAddress(), PROTOCOL_JOIN)
		}
	}
	return nil
}

// Send sends the message through the given Shoset network.
func (cjh *ConfigJoinHandler) Send(s *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Send not implemented")
}

// Wait returns the message received for a given Shoset.
func (cjh *ConfigJoinHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigJoinHandler.Wait not implemented")
	return nil
}
