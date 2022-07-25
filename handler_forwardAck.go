package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
)

// forwardAcknownledgeHandler implements MessageHandlers interface.
type forwardAcknownledgeHandler struct{}

// Get returns the message for a given ShosetConn.
func (fah *forwardAcknownledgeHandler) Get(c *ShosetConn) (msg.Message, error) {
	var m msg.ForwardAck
	err := c.GetReader().ReadMessage(&m)
	return m, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (fah *forwardAcknownledgeHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	//fmt.Println("((fah *forwardAcknownledgeHandler) HandleDoubleWay) Lname : ", c.GetLocalLogicalName(), " message : ", message)

	m := message.(msg.ForwardAck)
	if notInQueue := c.GetShoset().Queue["forwardAck"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle forwardAck")
	}

	c.GetShoset().MessageEventBus.Publish("forwardAck", true) // Sent data is not used

	return nil
}

// Send sends the message through the given Shoset network.
func (fah *forwardAcknownledgeHandler) Send(s *Shoset, m msg.Message) {
	//s.forwardMessage(m)
}

// Wait returns the message received for a given Shoset.
func (fah *forwardAcknownledgeHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

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
	s.MessageEventBus.Subscribe("forwardAck", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("forwardAck", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait forwardAck (timeout)")
			return nil
		case <-chNewMessage:
			//Check message presence in two steps to avoid accessing attributs of <nil>
			cell := replies.Get()
			if cell == nil {
				//time.Sleep(time.Duration(10) * time.Millisecond)
				break
			}
			message := cell.GetMessage()
			if message == nil {
				//time.Sleep(time.Duration(10) * time.Millisecond)
				break
			}
			forwardAck := message.(msg.ForwardAck)
			if forwardAck.OGMessageUUID == args["UUID"] {
				return &message
			}
		}
	}
}
