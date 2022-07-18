package shoset

import (
	"errors"
	"fmt"
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
	fmt.Println("((smh *SimpleMessageHandler) HandleDoubleWay) Lname : ", c.GetLocalLogicalName(), " message : ", message)
	m := message.(msg.SimpleMessage)
	if notInQueue := c.GetShoset().Queue["simpleMessage"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle simpleMessage")
	}

	// Notify of the new message
	select {
	case c.GetShoset().NewMessageEvent["simpleMessage"] <- true:
	default:
		fmt.Println("Nobody is waiting for NewMessageEvent[simpleMessage]")
	}

	return nil
}

// Send sends the message through the given Shoset network.
func (smh *SimpleMessageHandler) Send(s *Shoset, m msg.Message) {
	s.forwardMessage(m)
}

// Wait returns the message received for a given Shoset.
func (smh *SimpleMessageHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	for {
		// Possibilité de na pas voir passer un message
		select {
		case <-s.NewMessageEvent["simpleMessage"]:
			//Check message presence in two steps to avoid accessing attributs of <nil>
			cell := replies.Get()
			if cell == nil {
				break
			}
			message := cell.GetMessage()
			if message == nil {
				break
			}
			return &message

		case <-time.After(time.Duration(timeout) * time.Second):
			s.Logger.Warn().Msg("No message received in Wait : timeout")
			return nil
		}
	}
}
