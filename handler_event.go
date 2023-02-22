package shoset

import (
	"time"

	"github.com/ditrit/shoset/msg"
)

// EventHandler implements MessageHandlers interface.
type EventHandler struct{}

// Get returns the message for a given ShosetConn.
func (eh *EventHandler) Get(c *ShosetConn) (msg.Message, error) {
	var evt msg.Event
	err := c.GetReader().ReadMessage(&evt)
	return evt, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (eh *EventHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	evt := message.(msg.Event)
	if notInQueue := c.GetShoset().Queue["evt"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); notInQueue {
		eh.Send(c.GetShoset(), evt)
	}

	c.GetShoset().MessageEventBus.Publish("evt", true) // Notifies of the reception of a new message

	return nil
}

// Send sends the message through the given Shoset network.
func (eh *EventHandler) Send(s *Shoset, evt msg.Message) {
	_ = s.Queue["evt"].Push(evt, VOID, VOID)
	s.ConnsByLname.Iterate(
		func(lname string, ipAddress string, conn interface{}) {
			err := conn.(*ShosetConn).GetWriter().SendMessage(evt)
			if err != nil {
				s.Logger.Warn().Msg(s.GetLogicalName() + " couldn't send evt to " + conn.(*ShosetConn).GetRemoteLogicalName() + ": " + err.Error())
			} else {
				s.Logger.Debug().Msg("evt sent successfully from " + s.GetLogicalName() + " to " + conn.(*ShosetConn).GetRemoteLogicalName())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (eh *EventHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	topicName, ok := args["topic"]
	if !ok {
		s.Logger.Error().Msg("no topic provided for Wait evt")
		return nil
	}

	// Checks every message in the queue before waiting for a new message
	// Checks message presence in two steps to avoid accessing attributs of <nil>
	for {
		cell := replies.Get()
		if cell != nil {
			message := cell.GetMessage()
			if message != nil {
				event := message.(msg.Event)
				if event.GetTopic() == topicName && (args["event"] == VOID || event.GetEvent() == args["event"]) {
					return &message
				}
			}
		} else {
			// Locking Queue to avoid missing a message while preparing the channel to receive events.
			replies.GetQueue().LockQueue()
			break
		}
	}
	// Creating channel
	chNewMessage := make(chan interface{})
	s.MessageEventBus.Subscribe("evt", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("evt", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait event (timeout).")
			return nil
		case <-chNewMessage:
			//Checks message presence in two steps to avoid accessing fields of <nil>
			s.Logger.Debug().Msg("New message received in Wait event.")
			cell := replies.Get()
			if cell == nil {
				break
			}
			message := cell.GetMessage()
			if message == nil {
				break
			}
			event := message.(msg.Event)
			if event.GetTopic() == topicName && (args["event"] == VOID || event.GetEvent() == args["event"]) {
				return &message
			}
		}
	}
}
