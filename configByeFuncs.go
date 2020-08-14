package shoset

import (
	"fmt"

	"github.com/ditrit/shoset/msg"
)

/*
	Addendum explicatif :
	Il est probablement intéressant de comparer l'état du projet à ce moment et avec la
	dernière contribution de Marie-Hélène ; mes quelques modifications sont suppposées être
	des corrections de bug, mais cela peut également être du à une mésinterpétation de ma part
	du fonctionnnement du code
	J'ai également fait un git rebase de cette branche avec Master afin que son architecture
	etc... soit à jour

	Les deux plus gros changements sont la liste utilisée dans le cas du bye (ConnsByAddr
	au lieu de ConnsJoin, qui faisait notemment planter le programme avec les clusters) et
	une autre petit modification sur l'enregistrement de nouvelles ShosetConn qui enregistrait
	le nom local au lieu du nom distant

	Actuellement, le problème principal que j'ai identifié est la réponse au bye (bye_ok donc)
	qui semble se perdre, avec la shoset voulant partir qui attend indéfiniment le bye_ok
	A noter que en lançant plusieurs fois le test, il est arrivé dans certains cas qu'un bye_ok
	arrive effectivement à bout et que la shoset sur le départ raccourcice effectivement la liste
	des connections en attente
*/

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
	fmt.Println(cfg.GetCommandName())
	switch cfg.GetCommandName() {
	case "bye":
		// received by instances connected to the shutting down
		// from the instance that is shutting down

		// remove connection from lists
		conn := ch.ConnsByAddr.m[msgSource]

		//
		defer ch.deleteConn(msgSource)

		/*
			The nex block was already comented out
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

		*/

		cfgByeOk := msg.NewCfgByeOk(ch.GetBindAddr())
		conn.SendMessage(cfgByeOk)

	case "bye_ok":
		// received by the instance that is shutting down,
		// from the instances that received the Bye msg
		ch.ConnsBye.Delete(msgSource)
		c.socket.Close()
	case "bye_bro":
		// received by brothers of the instance that is shutting down,
		// from another connection
		ch.NameBrothers.Delete(msgSource)
	}
	return nil
}
