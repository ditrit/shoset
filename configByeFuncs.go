package shoset

import (
	"fmt"

	"github.com/ditrit/shoset/msg"
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
	fmt.Printf("$$$$$$$$$$$ You're in shoset : %v\n", ch.lName)
	// fmt.Printf("%v : %s",ch.lName, ch.String())
	fmt.Println(cfg.GetCommandName())
	switch cfg.GetCommandName() {
	case "bye":
		// Does not work for aggregator as they don't have ConnsJoin

		for connAddr, conn := range ch.ConnsByAddr.m {
			fmt.Printf("%v : %v\n", connAddr, conn)
		}

		// received by instances connected to the shutting down
		// from the instance that is shutting down

		// remove connection from lists
		// fmt.Printf("ConnsByAddr : %v\n", ch.ConnsByAddr.m)
		fmt.Printf("========msgSource : %v\n", msgSource)
		// fmt.Printf("========ch.ConnsByAddr.m[msgSource]: %v\n", ch.ConnsByAddr.m)
		// for addr, coon := range ch.ConnsByAddr.m {
		// 	fmt.Printf("%v  :  %v\n", addr, coon)
		// }
		conn, err := ch.ConnsByAddr.m[msgSource]
		fmt.Printf("status: %v\n", err)
		// fmt.Printf("conn : %s\n", conn.String())
		defer ch.deleteConn(msgSource)
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

		//Trying to fix the test
		//fmt.Printf("conn : %s\n", conn.String())

		cfgByeOk := msg.NewCfgByeOk(ch.GetBindAddr())
		fmt.Printf("=======> sending bye ok from socket %v to socket %v\n", ch.lName, c.name)
		conn.SendMessage(cfgByeOk)

	case "bye_ok":
		// received by the instance that is shutting down,
		// from the instances that received the Bye msg
		fmt.Printf("&&&&&& Receiver bye_ok from %v %v %v\n", c.name, c.bindAddr, c.addr)
		ch.ConnsBye.Delete(msgSource)
		fmt.Printf("ConnsBye after removing %v : %v\n", msgSource, ch.ConnsBye.m)
		fmt.Printf("closing socket\n")
		c.socket.Close()
	case "bye_bro":
		// received by brothers of the instance that is shutting down,
		// from another connection
		ch.NameBrothers.Delete(msgSource)
	}
	return nil
}
