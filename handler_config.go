package shoset

import (
	"errors"
	"time"

	"github.com/ditrit/shoset/msg"
)

// ConfigHandler implements MessageHandlers interface.
type ConfigHandler struct{}

// MOVE TO GANDALF
// Get returns the message for a given ShosetConn.
func (ch *ConfigHandler) Get(c *ShosetConn) (msg.Message, error) {
	var conf msg.Config
	err := c.GetReader().ReadMessage(&conf)
	return conf, err
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (ch *ConfigHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	conf := message.(msg.Config)
	if notInQueue := c.GetShoset().Queue["cmd"].Push(conf, c.GetRemoteShosetType(), c.GetLocalAddress()); !notInQueue {
		return errors.New("failed to handle command")
	}
	return nil
}

// Send sends the message through the given Shoset network.
func (ch *ConfigHandler) Send(s *Shoset, cmd msg.Message) {
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			if err := conn.(*ShosetConn).GetWriter().SendMessage(cmd); err != nil {
				conn.(*ShosetConn).Logger.Warn().Msg("couldn't send config msg : " + err.Error())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (ch *ConfigHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
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
