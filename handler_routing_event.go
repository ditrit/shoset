package shoset

import (
	"fmt"

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
	routingEvt := message.(msg.RoutingEvent)

	originLogicalName := routingEvt.GetOrigin()

	value, ok := c.GetShoset().RouteTable.Load(originLogicalName)

	shosetLname := c.GetLocalLogicalName()

	if c.GetLocalLogicalName() == originLogicalName {
		// There are no shosetConn to self or brothers
		c.GetShoset().RouteTable.Store(originLogicalName, NewRoute(c.GetRemoteLogicalName(), c, 2, routingEvt.GetUUID(), routingEvt.Timestamp))
		return nil
	} else if ok {
		if (value.(Route).GetUUID() != routingEvt.GetUUID() && routingEvt.Timestamp > value.(Route).timestamp) || (routingEvt.GetNbSteps() < value.(Route).nb_steps) { //UUID is different if Route is invalid and need to be replaced
			// Save route
			fmt.Printf("\n(HandleDoubleWay) shosetLname : %v \n\t message : %v \n\t value : %v ok : %v \nSave better Route.\n", shosetLname, message, value, ok)
			
			//c.GetShoset().RouteTable.Delete(originLogicalName)
			c.GetShoset().RouteTable.Store(originLogicalName, NewRoute(c.GetRemoteLogicalName(), c, routingEvt.GetNbSteps(), routingEvt.GetUUID(), routingEvt.Timestamp))

			// Send NewRouteEvent
			select {
			case c.GetShoset().NewRouteEvent <- originLogicalName:
				fmt.Println("Sending NewRouteEvent")
			default:
				fmt.Println("Nobody is waiting for NewRouteEvent")
			}

			// Rebroadcast Routing event
			routingEvt.SetNbSteps(routingEvt.GetNbSteps() + 1)
			reh.Send(c.GetShoset(), routingEvt)
			return nil
		} else {
			// Route not worse saving
			fmt.Printf("\n(HandleDoubleWay) shosetLname : %v \n\t message : %v \n\t value : %v ok : %v \nRoute not worse saving.\n", shosetLname, message, value, ok)
			return nil
		}
	}
	// Unknown Route (Origin not found in RouteTable)

	c.GetShoset().RouteTable.Store(originLogicalName, NewRoute(c.GetRemoteLogicalName(), c, routingEvt.GetNbSteps(), routingEvt.GetUUID(), routingEvt.Timestamp))

	// Send NewRouteEvent
	select {
	case c.GetShoset().NewRouteEvent <- originLogicalName:
		fmt.Println("Sending NewRouteEvent")
	default:
		fmt.Println("Nobody is waiting for NewRouteEvent")
	}

	// Reoute trigered every time the route is unknown :

	fmt.Printf("\n(HandleDoubleWay) shosetLname : %v \n\t message : %v \n\t value : %v ok : %v \nStore unknown Route.\n", shosetLname, message, value, ok)

	reRouting := msg.NewRoutingEvent(c.GetLocalLogicalName(), routingEvt.GetUUID())
	reh.Send(c.GetShoset(), reRouting)
	//fmt.Println("Reroute : ", c.GetLocalLogicalName())

	// Rebroadcast Routing event
	routingEvt.SetNbSteps(routingEvt.GetNbSteps() + 1)
	reh.Send(c.GetShoset(), routingEvt)

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
