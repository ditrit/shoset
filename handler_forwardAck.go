package shoset

import (
	"errors"
	"fmt"
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
	fmt.Println("((fah *forwardAcknownledgeHandler) HandleDoubleWay) Lname : ", c.GetLocalLogicalName(), " message : ", message)

	m := message.(msg.ForwardAck)
	if notInQueue := c.GetShoset().Queue["forwardAck"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle forwardAck")
	}

	return nil
}

// Send sends the message through the given Shoset network.
func (fah *forwardAcknownledgeHandler) Send(s *Shoset, m msg.Message) {
	//s.forwardMessage(m)
}

// Wait returns the message received for a given Shoset.
func (fah *forwardAcknownledgeHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait forwardAck (timeout)")
			return nil
		default:
			//Check message presence in two steps to avoid accessing attributs of <nil>
			cell := replies.Get()
			if cell == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			message := cell.GetMessage()
			if message == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}

			forwardAck := message.(msg.ForwardAck)
			if forwardAck.OGMessageUUID == args["UUID"] {
				return &message
			}

		}

	}
}
