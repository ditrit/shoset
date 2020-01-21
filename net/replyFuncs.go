package net

import (
	"fmt"
	"time"

	"../msg"
)

// HandleReply :
func HandleReply(c *ShosetConn) error {
	var rep msg.Reply
	err := c.ReadMessage(&rep)
	c.GetCh().FQueue("rep").Push(rep)
	return err
}

// SendReplyConn :
func SendReplyConn(c *ShosetConn, rep interface{}) {
	fmt.Print("Sending reply.\n")
	c.WriteString("rep")
	c.WriteMessage(rep)
}

// SendReply :
func SendReply(c *Shoset, rep msg.Message) {
	fmt.Print("Sending Reply.\n")
	for _, conn := range c.GetConnsByAddr() {
		conn.SendMessage(rep)
	}
}

// WaitReply :
func WaitReply(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	commandUUID, ok := args["uuid"]
	if !ok {
		return nil
	}
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				reply := (*message).(msg.Reply)
				if reply.GetCmdUUID() == commandUUID {
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
