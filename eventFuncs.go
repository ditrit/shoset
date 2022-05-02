package shoset

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ditrit/shoset/msg"
)

// GetEvent :
func GetEvent(c *ShosetConn) (msg.Message, error) {
	var evt msg.Event
	err := c.ReadMessage(&evt)
	return evt, err
}

// HandleEvent :
func HandleEvent(c *ShosetConn, message msg.Message) error {
	evt := message.(msg.Event)
	if state := c.GetCh().Queue["evt"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
		SendEvent(c.ch, evt)
	}
	return nil
}

// SendEventConn :
func SendEventConn(c *ShosetConn, evt interface{}) {
	_, err := c.WriteString("evt")
	if err != nil {
		log.Error().Msg("couldn't write string evt : " + err.Error())
		return
	}
	err = c.WriteMessage(evt)
	if err != nil {
		log.Error().Msg("couldn't write message evt : " + err.Error())
		return
	}
}

// SendEvent :
func SendEvent(c *Shoset, evt msg.Message) {
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			err := conn.SendMessage(evt)
			if err != nil {
				log.Error().Msg("couldn't send evt : " + err.Error())
			}
		},
	)
}

// WaitEvent :
func WaitEvent(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
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
			if message != nil {
				event := message.(msg.Event)
				if event.GetTopic() == topicName && (eventName == "" || event.GetEvent() == eventName) {
					term <- &message
				}
			} else {
				time.Sleep(time.Duration(10) * time.Millisecond)
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
