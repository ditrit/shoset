package shoset

import (
	"sync"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// ConfigByeHandler implements MessageHandlers interface.
type ConfigByeHandler struct{}

// Get returns the message for a given ShosetConn.
func (cbh *ConfigByeHandler) Get(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.GetReader().ReadMessage(&cfg)
	return cfg, err

}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (cbh *ConfigByeHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol)

	// Finds and deletes Routes going through the Lname initiating bye/delete (not tester)
	// Chercher le Lname ou juste la Conn ?
	deleteLname := c.GetRemoteLogicalName()
	c.GetShoset().RouteTable.Range(
		func(key, val interface{}) bool {
			if val.(Route).neighbour == deleteLname {
				c.GetShoset().RouteTable.Delete(key)
			}
			return true
		})

	switch cfg.GetCommandName() {
	case PROTOCOL_EXIT:
		// incoming bye request.
		// send delete signal to all connected shosets from our list of known shosets.
		cfgNewDelete := msg.NewConfigProtocol(cfg.GetAddress(), c.GetShoset().GetLogicalName(), c.GetShoset().GetShosetType(), DELETE)
		c.GetShoset().ConnsByLname.Iterate(
			func(address string, bro interface{}) {
				if address != cfg.GetAddress() {
					if err := bro.(*ShosetConn).GetWriter().SendMessage(*cfgNewDelete); err != nil {
						bro.(*ShosetConn).Logger.Warn().Msg("couldn't send cfgNewDelete : " + err.Error())
					}
				}
			},
		)

	case DELETE:
		// incoming delete signal.
		// forget the concerned shoset from our list of known shosets and close connection.
		mapSync := new(sync.Map)
		mapSync.Store(cfg.GetLogicalName(), true)
		c.GetShoset().LnamesByProtocol.Store(PROTOCOL_EXIT, mapSync)
		c.GetShoset().LnamesByType.Store(cfg.GetShosetType(), mapSync)
		c.GetShoset().deleteConn(cfg.GetAddress(), cfg.GetLogicalName())
	}
	return nil
}

// Send sends the message through the given Shoset network.
func (cbh *ConfigByeHandler) Send(s *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ConfigByeHandler.Send not implemented")
}

// Wait returns the message received for a given Shoset.
func (cbh *ConfigByeHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("ConfigByeHandler.Wait not implemented")
	return nil
}
