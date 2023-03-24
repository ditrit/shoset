package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// ForwardAcknownledgeHandler implements MessageHandlers interface.
type ForwardAcknownledgeHandler struct{}

// Get returns the message for a given ShosetConn.
func (fah *ForwardAcknownledgeHandler) Get(c *ShosetConn) (msg.Message, error) {
	var m msg.ForwardAck
	err := c.GetReader().ReadMessage(&m)
	return m, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (fah *ForwardAcknownledgeHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	m := message.(msg.ForwardAck)
	if notInQueue := c.GetShoset().Queue["forwardAck"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle forwardAck")
	}

	c.GetShoset().MessageEventBus.Publish("forwardAck", true) // Notifies of the reception of a new message

	return nil
}

// Send sends the message through the given Shoset network.
func (fah *ForwardAcknownledgeHandler) Send(s *Shoset, m msg.Message) {
	// no-op
	log.Warn().Msg("ForwardAcknownledgeHandler.Send not implemented")
}

// Wait returns the message received for a given Shoset.
func (fah *ForwardAcknownledgeHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	// Checks every message in the queue before waiting for a new message
	//Checks message presence in two steps to avoid accessing attributs of <nil>
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
	s.MessageEventBus.Subscribe("forwardAck", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("forwardAck", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait ForwardAck (timeout).")
			return nil
		case <-chNewMessage:
			//Checks message presence in two steps to avoid accessing attributs of <nil>
			cell := replies.Get()
			if cell == nil {
				break
			}
			message := cell.GetMessage()
			if message == nil {
				break
			}
			forwardAck := message.(msg.ForwardAck)
			// Checks that it is a forwardAck for the right message.
			if forwardAck.OGMessageUUID == args["UUID"] {
				return &message
			}
		}
	}
}
