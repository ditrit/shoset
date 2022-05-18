package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
)

type ConfigHandler struct{}

// MOVE TO GANDALF
// GetConfig :
func (ch *ConfigHandler) Get(c *ShosetConn) (msg.Message, error) {
	var conf msg.Config
	err := c.ReadMessage(&conf)
	return conf, err
}

// HandleConfig :config
func (ch *ConfigHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	conf := message.(msg.Config)
	if !c.GetCh().Queue["cmd"].Push(conf, c.GetRemoteShosetType(), c.GetLocalAddress()) {
		return errors.New("failed to handle command")
	}
	return nil
}

// SendConfig :
func (ch *ConfigHandler) Send(c *Shoset, cmd msg.Message) {
	c.ConnsByName.Iterate(
		func(key string, conn interface{}) {
			if err := conn.(*ShosetConn).SendMessage(cmd); err != nil {
				conn.(*ShosetConn).logger.Warn().Msg("couldn't send config msg : " + err.Error())
			}
		},
	)
}

// WaitConfig :
func (ch *ConfigHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
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
			config := message.(msg.Config)
			if config.GetCommand() == commandName {
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
