package net

import (
	"fmt"
	"time"

	"../msg"
)

// HandleEvent :
func HandleEvent(c *ShosetConn) error {
	var evt msg.Event
	err := c.ReadMessage(&evt)
	c.GetCh().FQueue("evt").Push(evt, c.ShosetType, c.bindAddr)
	return err
}

// SendEventConn :
func SendEventConn(c *ShosetConn, evt interface{}) {
	fmt.Print("Sending config.\n")
	c.WriteString("evt")
	c.WriteMessage(evt)
}

// SendEvent :
func SendEvent(c *Shoset, evt msg.Message) {
	fmt.Print("Sending event.\n")
	for _, conn := range c.GetConnsByAddr() {
		conn.SendMessage(evt)
	}
}

// WaitEvent :
func WaitEvent(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	topicName, ok := args["topic"]
	if !ok {
		return nil
	}
	eventName := args["event"]
	term := make(chan *msg.Message, 1)
	cont := true
	go func() {
		for cont {
			message := replies.Get()
			if message != nil {
				event := (*message).(msg.Event)
				if event.GetTopic() == topicName && (eventName == "" || event.GetEvent() == eventName) {
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
