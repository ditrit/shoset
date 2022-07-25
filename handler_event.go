package shoset

import (
	"time"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
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

	c.GetShoset().MessageEventBus.Publish("evt", true) // Sent data is not use

	return nil
}

// Send sends the message through the given Shoset network.
func (eh *EventHandler) Send(s *Shoset, evt msg.Message) {
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			//conn.(*ShosetConn).WaitForValid() //
			err := conn.(*ShosetConn).GetWriter().SendMessage(evt)
			if err != nil {
				log.Warn().Msg("couldn't send evt : " + err.Error())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (eh *EventHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	topicName, ok := args["topic"]
	if !ok {
		s.Logger.Warn().Msg("no topic provided for Wait evt")
		return nil
	}

	// Check every message in the queue before waiting for a new message
	//Check message presence in two steps to avoid accessing attributs of <nil>
	for {
		cell := replies.Get()
		if cell != nil {
			message := cell.GetMessage()
			if message != nil {
				return &message
			}
		} else {
			replies.GetQueue().LockQueue()
			break			
		}
	}
	// Creation channel
	chNewMessage := make(chan interface{})

	// Inscription channel
	s.MessageEventBus.Subscribe("evt", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("evt", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait evt (timeout)")
			return nil
		case <-chNewMessage:
			//Check message presence in two steps to avoid accessing attributs of <nil>
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
