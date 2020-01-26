package net

import (
	"fmt"
	"time"

	"shoset/msg"
)

// HandleCommand :
func HandleCommand(c *ShosetConn) error {
	var cmd msg.Command
	err := c.ReadMessage(&cmd)
	c.GetCh().FQueue("cmd").Push(cmd, c.ShosetType, c.bindAddr)
	return err
}

// SendCommandConn :
func SendCommandConn(c *ShosetConn, cmd interface{}) {
	fmt.Print("Sending config.\n")
	c.WriteString("cmd")
	c.WriteMessage(cmd)
}

// SendCommand :
func SendCommand(c *Shoset, cmd msg.Message) {
	fmt.Print("Sending Command.\n")
	for _, conn := range c.GetConnsByAddr() {
		conn.SendMessage(cmd)
	}
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
	case res := <- term:
		cont = false
		return res
	case <-time.After(time.Duration(timeout) * time.Second):
		return nil
	}
}
