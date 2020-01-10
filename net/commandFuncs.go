package net

import (
	"fmt"
	"time"

	"../msg"
)

// HandleCommand :
func HandleCommand(c *ChaussetteConn) error {
	var cmd msg.Command
	err := c.ReadMessage(&cmd)
	c.GetCh().FQueue("cmd").Push(cmd)
	return err
}

// SendCommandConn :
func SendCommandConn(c *ChaussetteConn, cmd interface{}) {
	fmt.Print("Sending config.\n")
	c.WriteString("cmd")
	c.WriteMessage(cmd)
}

// SendCommand :
func SendCommand(c *Chaussette, cmd interface{}) {
	fmt.Print("Sending Command.\n")
	for _, conn := range c.GetConnsByAddr() {
		SendCommandConn(conn, cmd)
	}
}

// WaitCommand :
func WaitCommand(c *Chaussette, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	commandName, ok := args["name"]
	if !ok {
		return nil
	}
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				command := (*message).(msg.Command)
				if command.GetCommand() == commandName {
					term <- message
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
