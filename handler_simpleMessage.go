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
	return nil
}

// Send sends the message through the given Shoset network.
// Code duplicated in ShosetConn.go (handleMessageType)
func (smh *SimpleMessageHandler) Send(s *Shoset, m msg.Message) {
	// Forward the message using the RouteTable to get the next destination
	route, ok := s.RouteTable.Load(m.GetDestinationLname())
	if !ok {
		s.Logger.Error().Msg("Forward message : Failed to forward message destined to " + m.GetDestinationLname() + " No Route.")
	} else {
		fmt.Println("((SimpleMessageHandler) Send) ", s.GetLogicalName(), " is sending a message to ", m.GetDestinationLname(), "through ", route.(Route).neighbour, ".")
		
		
		err := route.(Route).GetNeighborConn().GetWriter().SendMessage(m)
		if err != nil {
			s.Logger.Error().Msg("Couldn't send forwarded message : " + err.Error())
		}
	}
}

// Wait returns the message received for a given Shoset.
func (smh *SimpleMessageHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
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
			term <- &message
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
