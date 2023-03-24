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
	timer := time.NewTimer(time.Duration(timeout) * time.Second)

	commandName, ok := args["name"]
	if !ok {
		s.Logger.Error().Msg("no command name provided for Wait")
		return nil
	}

	// Checks every message in the queue before waiting for a new message
	// Checks message presence in two steps to avoid accessing attributs of <nil>
	for {
		cell := replies.Get()
		if cell != nil {
			message := cell.GetMessage()
			if message != nil {
				cmd := message.(msg.Command)
				if cmd.GetCommand() == commandName {
					return &message
				}
			}
		} else {
			// Locking Queue to avoid missing a message while preparing the channel to receive events.
			replies.GetQueue().LockQueue()
			break
		}
	}
	// Creating channel
	chNewMessage := make(chan interface{})
	s.MessageEventBus.Subscribe("cmd", chNewMessage)
	replies.GetQueue().UnlockQueue()
	defer s.MessageEventBus.UnSubscribe("cmd", chNewMessage)

	for {
		select {
		case <-timer.C:
			s.Logger.Warn().Msg("No message received in Wait Command (timeout).")
			return nil
		case <-chNewMessage:
			//Checks message presence in two steps to avoid accessing fields of <nil>
			s.Logger.Debug().Msg("New message received in Wait Command.")
			cell := replies.Get()
			if cell == nil {
				break
			}
			message := cell.GetMessage()
			if message == nil {
				break
			}
			cmd := message.(msg.Command)
			if cmd.GetCommand() == commandName {
				return &message
			}
		}
	}
}
