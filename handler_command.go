package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
)

type CommandHandler struct{}

// GetCommand :
func (ch *CommandHandler) Get(c *ShosetConn) (msg.Message, error) {
	var m msg.Command
	err := c.ReadMessage(&m)
	return m, err
}

// HandleCommand :
func (ch *CommandHandler) Handle(c *ShosetConn, message msg.Message) error {
	m := message.(msg.Command)
	if !c.GetCh().Queue["cmd"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()) {
		return errors.New("failed to handle command")
	}
	return nil
}

// SendCommand :
func (ch *CommandHandler) Send(c *Shoset, m msg.Message) {
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			if err := conn.SendMessage(m); err != nil {
			}
		},
	)
}

// WaitCommand :
func (ch *CommandHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
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
