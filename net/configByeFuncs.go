package net

import (
	"fmt"
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
	fmt.Printf("msgSource : %v\n", msgSource)
	ch := c.GetCh()
	switch cfg.GetCommandName() {
	case "bye":
		// received by instances connected to the shutting down
		// from the instance that is shutting down

		// remove connection from lists
		fmt.Printf("ConnsJoin : %v\n", ch.ConnsJoin.m)
		conn := ch.ConnsJoin.m[msgSource]
		ch.deleteConn(msgSource)
		// // send RemoveBro to the brothers of the shutting down
		// msgRemoveBro := msg.NewCfgByeBro(msgSource)
		// fmt.Printf("cfg.LogicalName : %v\n", cfg.LogicalName)
		// fmt.Printf("ConnsByName : %v\n", ch.ConnsByName.m)
		// brothers := ch.ConnsByName.m[cfg.LogicalName]
		// fmt.Printf("brothers : %v\n", brothers.m)
		// if len(brothers.m) > 0 {
		// 	for _, conn := range brothers.m {
		// 		conn.SendMessage(msgRemoveBro)
		// 	}
		// }
		// send ack
		fmt.Printf("conn : %s\n", conn.String())
		cfgByeOk := msg.NewCfgByeOk(ch.GetBindAddr())
		fmt.Printf("sending bye ok\n")
		conn.SendMessage(cfgByeOk)
	case "bye_ok":
		// received by the instance that is shutting down,
		// from the instances that received the Bye msg
		ch.ConnsBye.Delete(msgSource)
		fmt.Printf("ConnsBye after removing %v : %v\n", msgSource, ch.ConnsBye.m)
		fmt.Printf("closing socket\n")
		c.socket.Close()
	case "bye_brother":
		// received by brothers of the instance that is shutting down,
		// from another connection
		ch.NameBrothers.Delete(msgSource)
	}
	return nil
}
