package shoset

import (
	"time"

	"github.com/ditrit/shoset/msg"
)

// GetCommand :
func GetCommand(c *ShosetConn) (msg.Message, error) {
	var cmd msg.Command
	err := c.ReadMessage(&cmd)
	return cmd, err
}

// HandleCommand :
func HandleCommand(c *ShosetConn, message msg.Message) error {
	cmd := message.(msg.Command)
	c.GetCh().Queue["cmd"].Push(cmd, c.GetRemoteShosetType(), c.GetLocalAddress())
	return nil
}

// SendCommand :
func SendCommand(c *Shoset, cmd msg.Message) {
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			err := conn.SendMessage(cmd)
			if err != nil {
			}
		},
	)
}

// WaitCommand :
func WaitCommand(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	commandName, ok := args["name"]
	if !ok {
		return nil
	}
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get().GetMessage()
			if message != nil {
				command := message.(msg.Command)
				if command.GetCommand() == commandName {
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
