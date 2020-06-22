package net

import (
	"shoset/msg"
)

// GetConfigBye :
func GetConfigBye(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigBye
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigBye :
func HandleConfigBye(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigBye)
	msgSource := cfg.GetBindAddress()
	ch := c.GetCh()
	switch cfg.GetCommandName() {
	case "bye":
		// remove connection from lists
		ch.deleteConn(msgSource)
		// send ack
		conn := ch.ConnsBye.m[msgSource]
		cfgByeOk := msg.NewCfgByeOk(msgSource)
		conn.SendMessage(cfgByeOk)
		// close connection
		c.socket.Close()
	case "bye_ok":
		ch.ConnsBye.Delete(msgSource)
		c.socket.Close()
	}
	return nil
}
