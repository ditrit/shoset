package shoset

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset/msg"
)

//TODO MOVE TO GANDALF
// GetConfig :
func GetConfig(c *ShosetConn) (msg.Message, error) {
	var conf msg.Config
	err := c.ReadMessage(&conf)
	return conf, err
}

// HandleConfig :config
func HandleConfig(c *ShosetConn, message msg.Message) error {
	conf := message.(msg.Config)
	c.GetCh().Queue["cmd"].Push(conf, c.GetRemoteShosetType(), c.GetLocalAddress())
	return nil
}

// SendConfig :
func SendConfig(c *Shoset, cmd msg.Message) {
	fmt.Print("Sending Config.\n")
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			err := conn.SendMessage(cmd)
			if err != nil {
				fmt.Println("couldn't send config msg", err)
			}
		},
	)
}

// WaitConfig :
func WaitConfig(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
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
				config := message.(msg.Config)
				if config.GetCommand() == commandName {
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
