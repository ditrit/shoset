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
	return nil
}

// Send sends the message through the given Shoset network.
func (eh *EventHandler) Send(s *Shoset, evt msg.Message) {
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			err := conn.(*ShosetConn).SendMessage(evt)
			if err != nil {
				log.Warn().Msg("couldn't send evt : " + err.Error())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (eh *EventHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	topicName, ok := args["topic"]
	if !ok {
		return nil
	}
	eventName := args["event"]
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get().GetMessage()
			if message == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			event := message.(msg.Event)
			if event.GetTopic() == topicName && (eventName == VOID || event.GetEvent() == eventName) {
				term <- &message
			}
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
