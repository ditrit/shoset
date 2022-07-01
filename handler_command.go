package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
)

// CommandHandler implements MessageHandlers interface.
type CommandHandler struct{}

// Get returns the message for a given ShosetConn.
func (ch *CommandHandler) Get(c *ShosetConn) (msg.Message, error) {
	var m msg.Command
	err := c.GetReader().ReadMessage(&m)
	return m, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (ch *CommandHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	m := message.(msg.Command)
	if notInQueue := c.GetShoset().Queue["cmd"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle command")
	}
	return nil
}

// Send sends the message through the given Shoset network.
func (ch *CommandHandler) Send(s *Shoset, m msg.Message) {
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			if err := conn.(*ShosetConn).GetWriter().SendMessage(m); err != nil {
				conn.(*ShosetConn).Logger.Warn().Msg("couldn't send command msg : " + err.Error())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (ch *CommandHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	commandName, ok := args["name"]
	if !ok {
		return nil
	}
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get().GetMessage()
			if message == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			command := message.(msg.Command)
			if command.GetCommand() == commandName {
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
