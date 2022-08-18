package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
)

// SimpleMessageHandler implements MessageHandlers interface.
type SimpleMessageHandler struct{}

// Get returns the message for a given ShosetConn.
func (smh *SimpleMessageHandler) Get(c *ShosetConn) (msg.Message, error) {
	var m msg.SimpleMessage
	err := c.GetReader().ReadMessage(&m)
	return m, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (smh *SimpleMessageHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	m := message.(msg.SimpleMessage)

	if notInQueue := c.GetShoset().Queue["simpleMessage"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle simpleMessage")
	}

	c.GetShoset().MessageEventBus.Publish("simpleMessage", true) // Notifies of the reception of a new message

	return nil
}

// Send sends the message through the given Shoset network.
func (smh *SimpleMessageHandler) Send(s *Shoset, m msg.Message) {
	s.forwardMessage(m)
}

// Wait returns the message received for a given Shoset.
func (smh *SimpleMessageHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	// Checks every message in the queue before waiting for a new message
	// Checks message presence in two steps to avoid accessing attributs of <nil>
	for {
		cell := replies.Get()
		if cell != nil {
			message := cell.GetMessage()
			if message != nil {
				return &message
			}
		} else {
			// Locking Queue to avoid missing a message while preparing the channel to receive events.
			replies.GetQueue().LockQueue()
			break
		}
	}
	// Creating channel
	chNewMessage := make(chan interface{})
	s.MessageEventBus.Subscribe("simpleMessage", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("simpleMessage", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait SimpleMessage (timeout).")
			return nil
		case <-chNewMessage:
			//Checks message presence in two steps to avoid accessing fields of <nil>
			cell := replies.Get()
			if cell == nil {
				break
			}
			message := cell.GetMessage()
			if message == nil {
				break
			}
			return &message
		}
	}
}
