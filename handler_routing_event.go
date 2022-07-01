package shoset

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// RoutingEventHandler implements MessageHandlers interface.
type RoutingEventHandler struct{}

// Get returns the message for a given ShosetConn.
func (reh *RoutingEventHandler) Get(c *ShosetConn) (msg.Message, error) {
	var revt msg.RoutingEvent
	err := c.GetReader().ReadMessage(&revt)
	return revt, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (reh *RoutingEventHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	fmt.Println("((reh *RoutingEventHandler) HandleDoubleWay)",message)

	routingEvt := message.(msg.RoutingEvent)

	originLogicalName := routingEvt.GetOrigin()

	if c.GetLocalLogicalName() == originLogicalName {
		c.GetShoset().RouteTable.Store(originLogicalName, NewRouter(c.GetLocalLogicalName(), 1, routingEvt.GetUUID()))
		return nil
	} else if value, ok := c.GetShoset().RouteTable.Load(originLogicalName); ok {
		fmt.Println("((reh *RoutingEventHandler) HandleDoubleWay) value : ", value)
		if (value.(Route).GetUUID() != routingEvt.GetUUID()) || (routingEvt.GetNbSteps() < value.(Route).nb_steps) { //UUID is different if Route is invalid and need to be replaced
			// Save route
			c.GetShoset().RouteTable.Delete(originLogicalName)
		} else {
			return nil
		}
		// Route not worse saving
	}
	// Unknown Route (Origin not found in RouteTable)
	c.GetShoset().RouteTable.Store(originLogicalName, NewRouter(c.GetRemoteLogicalName(), routingEvt.GetNbSteps(), routingEvt.GetUUID()))

	// // Reroute when savin new Route
	// reRouting := msg.NewRoutingEvent(c.GetLocalLogicalName())
	// reh.Send(c.GetShoset(), reRouting)

	// Rebroadcast Routing event
	routingEvt.SetNbSteps(routingEvt.GetNbSteps() + 1)

	reh.Send(c.GetShoset(), routingEvt)

	return nil
}

// Send sends the message through the given Shoset network.
func (reh *RoutingEventHandler) Send(s *Shoset, evt msg.Message) { // Add send to logical name
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			err := conn.(*ShosetConn).SendMessage(evt)
			if err != nil {
				log.Warn().Msg("couldn't send evt : " + err.Error())
			}
		},
	)
}

// // Rebroadcast : Sends out RoutingEvent that was imporoving route
// func (reh *RoutingEventHandler) Rebroadcast(s *Shoset, r msg.RoutingEvent) {
// 	r.Nb_steps++

// 	reh.Send(s,r)
// }

// Wait returns the message received for a given Shoset.
func (reh *RoutingEventHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	//eventName := args["event"]

	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get().GetMessage()
			if message == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			event := message.(msg.RoutingEvent)

			//

			//if event.GetTopic() == topicName && (eventName == VOID || event.GetEvent() == eventName) {
			term <- &message
			//}
			fmt.Println("((RoutingEventHandler) Wait) : ", event)
		}
	}()
	select {
	case res := <-term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}
