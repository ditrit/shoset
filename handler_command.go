package shoset

import (
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
	// log the fact that we received a command
	c.Logger.Info().Msg("received command : " + m.GetPayload())
	if notInQueue := c.GetShoset().Queue["cmd"].Push(m, c.GetRemoteShosetType(), c.GetLocalAddress()); notInQueue {
		ch.Send(c.GetShoset(), m)
		c.GetShoset().MessageEventBus.Publish("cmd", true) // Notifies of the reception of a new message
	}

	c.GetShoset().MessageEventBus.Publish("cmd", true) // Notifies of the reception of a new message

	return nil
}

// Send sends the message through the given Shoset network. (through all the connections)
func (ch *CommandHandler) Send(s *Shoset, m msg.Message) {
	_ = s.Queue["cmd"].Push(m, VOID, VOID)
	// get the target of the message m
	var target = m.(msg.Command).GetTarget()
	s.ConnsByLname.Iterate(
		func(lname string, ipAddress string, conn interface{}) {
			if lname == target {
				s.Logger.Debug().Msg("sending command to " + ipAddress)
				if err := conn.(*ShosetConn).GetWriter().SendMessage(m); err != nil {
					conn.(*ShosetConn).Logger.Warn().Msg("couldn't send command msg : " + err.Error())
				}
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
			response := replies.Get()
			if response == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			message := response.GetMessage()
			if message == nil {
				time.Sleep(time.Duration(10) * time.Millisecond)
				continue
			}
			command := message.(msg.Command)
			if command.GetCommand() == commandName {
				term <- &message
				return
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
