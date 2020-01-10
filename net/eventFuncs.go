package net

import (
	"fmt"
	"time"

	"../msg"
)

// HandleEvent :
func HandleEvent(c *ChaussetteConn) error {
	var evt msg.Event
	err := c.ReadMessage(&evt)
	c.GetCh().FQueue("evt").Push(evt)
	return err
}

// SendEventConn :
func SendEventConn(c *ChaussetteConn, evt interface{}) {
	fmt.Print("Sending config.\n")
	c.WriteString("evt")
	c.WriteMessage(evt)
}

// SendEvent :
func SendEvent(c *Chaussette, evt interface{}) {
	fmt.Print("Sending event.\n")
	for _, conn := range c.GetConnsByAddr() {
		SendEventConn(conn, evt)
	}
}

// WaitEvent :
func WaitEvent(c *Chaussette, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
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
