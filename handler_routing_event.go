package shoset

import (
	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// RoutingEventHandler implements MessageHandlers interface.
type RoutingEventHandler struct{}

// Get returns the message for a given ShosetConn.
func (reh *RoutingEventHandler) Get(c *ShosetConn) (msg.Message, error) {
	var routingEvt msg.RoutingEvent
	err := c.GetReader().ReadMessage(&routingEvt)
	return routingEvt, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (reh *RoutingEventHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	// Avoids using conns that are not yet stored and have missing informations (No remote Lname).
	if !c.GetIsValid() {
		return nil
	}

	routingEvt := message.(msg.RoutingEvent)

	originLogicalName := routingEvt.GetOrigin()

	// Tries to load the existing Route.
	value, ok := c.GetShoset().RouteTable.Load(originLogicalName)
	if ok && !((value.(Route).GetUUID() != routingEvt.GetUUID() && routingEvt.Timestamp > value.(Route).GetTimestamp()) || (routingEvt.GetNbSteps() < value.(Route).GetNbSteps())) { // See table in instructions.
		return nil
	}
	c.GetShoset().SaveRoute(c, &routingEvt)
	return nil
}

// Send sends the message through the given Shoset network.
func (reh *RoutingEventHandler) Send(s *Shoset, evt msg.Message) {
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			err := conn.(*ShosetConn).GetWriter().SendMessage(evt)
			if err != nil {
				log.Warn().Msg("couldn't send routingEvent : " + err.Error())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (reh *RoutingEventHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("RoutingEventHandler.Wait not implemented")
	return nil
}
