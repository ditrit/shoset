package main

import (
	"fmt"
	"time"

	"./msg"
)

// SendEvent : send an event
func (c *ChaussetteConn) SendEvent(evt *msg.Event) {
	c.wb.WriteString("evt")
	c.wb.WriteEvent(*evt)
}

// SendEvent : send an event...
// event is sent on each connection
func (c *Chaussette) SendEvent(evt *msg.Event) {
	fmt.Print("Sending event.\n")
	for _, conn := range c.connsByAddr {
		conn.SendEvent(evt)
	}
}

// SendCommand : Send a message
func (c *ChaussetteConn) SendCommand(cmd *msg.Command) {
	c.wb.WriteString("cmd")
	c.wb.WriteCommand(*cmd)
}

// SendCommand : Send a message
// todo : manage routing
//    identify relevant targets (routing info matches identity)
//    then try on each instance until success
func (c *Chaussette) SendCommand(cmd *msg.Command) {
	fmt.Print("Sending command.\n")
	for _, conn := range c.connsByAddr {
		conn.SendCommand(cmd)
	}
}

// SendReply :
func (c *ChaussetteConn) SendReply(rep *msg.Reply) {
	c.wb.WriteString("rep")
	c.wb.WriteReply(*rep)
}

// SendReply :
func (c *Chaussette) SendReply(rep *msg.Reply) {
	fmt.Print("Sending reply.\n")
	for _, conn := range c.connsByAddr {
		conn.SendReply(rep)
	}
}

// SendConfig :
func (c *ChaussetteConn) SendConfig(cfg *msg.Config) {
	fmt.Print("Sending config.\n")
	c.wb.WriteString("cfg")
	c.wb.WriteConfig(*cfg)
}

// SendConfig :
func (c *Chaussette) SendConfig(cfg *msg.Config) {
	fmt.Print("Sending configuration.\n")
	for _, conn := range c.connsByAddr {
		conn.SendConfig(cfg)
	}
}

// WaitEvent :
func (c *Chaussette) WaitEvent(events *msg.Iterator, topicName string, eventName string, timeout int) *msg.Event {
	term := make(chan *msg.Event, 1)
	cont := true
	go func() {
		for cont {
			message := events.Get()
			if message != nil {
				event := (*message).(msg.Event)
				if topicName == event.GetTopic() && eventName == event.GetEvent() {
					term <- &event
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

// WaitCommand : uniquement au sein d'un connecteur a priori
func (c *Chaussette) WaitCommand(commands *msg.Iterator, commandName string, timeout int) *msg.Command {
	term := make(chan *msg.Command, 1)
	cont := true
	go func() {
		for cont {
			message := commands.Get()
			if message != nil {
				command := (*message).(msg.Command)
				if commandName == command.GetCommand() {
					term <- &command
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

// WaitReply :
func (c *Chaussette) WaitReply(replies *msg.Iterator, commandUUID string, timeout int) *msg.Reply {
	term := make(chan *msg.Reply, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				reply := (*message).(msg.Reply)
				if commandUUID == reply.GetCmdUUID() {
					term <- &reply
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

// WaitConfig :
func (c *Chaussette) WaitConfig(replies *msg.Iterator, commandUUID string, timeout int) *msg.Config {
	term := make(chan *msg.Config, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				config := (*message).(msg.Config)
				term <- &config
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
