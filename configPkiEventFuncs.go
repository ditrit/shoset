package shoset

import (
	// "fmt"
	"io/ioutil"
	"os"

	"github.com/ditrit/shoset/msg"
)

// GetEvent :
func GetPkiEvent(c *ShosetConn) (msg.Message, error) {
	var evt msg.PkiEvent
	err := c.ReadMessage(&evt)
	return evt, err
}

// HandleEvent :
func HandlePkiEvent(c *ShosetConn, message msg.Message) error {

	evt := message.(msg.PkiEvent)
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// si je suis pki :
	//   si on m'envoie un certreq
	//   alors
	//     on extrait le certreq et le secret
	//     je signe
	//     je renvoie le resultat en precisant adresse dans un champ
	//     je reprend l'uuid du msg, je lui ajoute un caractere au bout (uuid_response) et je l'utilise comme uuid du msg de reponse
	//     return
	//   fi
	if c.ch.GetIsPki() {
		if evt.GetCertReq() != nil {
			// fmt.Println(c.ch.GetBindAddress(), "enters pki event")
			CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.crt")
			if err != nil {
				return err
			}

			signedCert := c.ch.SignCertificate(evt.GetCertReq(), evt.GetHostPublicKey())
			if signedCert != nil {
				var returnPkiEvent *msg.PkiEvent
				if c.ch.GetLogicalName() == evt.GetLogicalName() { // les clusters deviennent à leur tour pki
					CAprivateKey, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/privateCAKey.key")
					if err != nil {
						return err
					}
					returnPkiEvent = msg.NewPkiEventReturn(evt.GetRequestAddress(), signedCert, CAcert, CAprivateKey)
					returnPkiEvent.SetUUID(evt.GetUUID()+"*")
					// fmt.Println("return pki event sent to", evt.GetRequestAddress())
				} else {
					returnPkiEvent = msg.NewPkiEventReturn(evt.GetRequestAddress(), signedCert, CAcert, nil)
					returnPkiEvent.SetUUID(evt.GetUUID()+"*")
					// fmt.Println("return pki event sent to", evt.GetRequestAddress())
				}
				SendPkiEvent(c.ch, returnPkiEvent)
			}
		}
	} else if c.ch.GetBindAddress() == evt.GetRequestAddress() {
		// si le msg est une reponse a ma demmande (champ adresse equivaut la mienne), c'est donc moi qui ai envoyé le certreq
		// alors
		//   je recupere le msg et lire mon cert
		//   return
		// fi
		signedCert := evt.GetSignedCert()
		ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/cert.crt", signedCert, 0644)
		caCert := evt.GetCAcert()
		ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/CAcert.crt", caCert, 0644)

		if evt.GetCAprivateKey() != nil {
			caPrivateKey := evt.GetCAprivateKey()
			ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", caPrivateKey, 0644)

			c.ch.SetIsPki(true)
			// fmt.Println(c.ch.GetBindAddress(), "is now pki and has been certified")
		} else {
			// fmt.Println(c.ch.GetBindAddress(), "has been certified")
		}
	} else {
		// je transmet le msg puisque je suis ni pki ni demandeur
		if state := c.GetCh().Queue["pkievt"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
			SendPkiEvent(c.ch, evt)
			// fmt.Println(c.ch.GetBindAddress(), "has treated the event")
		}
	}
	return nil
}

// SendEventConn :
func SendPkiEventConn(c *ShosetConn, evt interface{}) {
	// fmt.Println("Sending pki config.\n")
	c.WriteString("pkievt")
	c.WriteMessage(evt)
}

// SendEvent :
func SendPkiEvent(c *Shoset, evt msg.Message) {
	// fmt.Println("Sending pki event.\n")
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			conn.SendMessage(evt)
		},
	)
}

// WaitEvent :
// func WaitPkiEvent(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
// 	topicName, ok := args["topic"]
// 	if !ok {
// 		return nil
// 	}
// 	eventName := args["event"]
// 	term := make(chan *msg.Message, 1)
// 	cont := true
// 	go func() {
// 		for cont {
// 			message := replies.Get().GetMessage()
// 			if message != nil {
// 				event := message.(msg.PkiEvent)
// 				if event.GetTopic() == topicName && (eventName == "" || event.GetEvent() == eventName) {
// 					term <- &message
// 				}
// 			} else {
// 				time.Sleep(time.Duration(10) * time.Millisecond)
// 			}
// 		}
// 	}()
// 	select {
// 	case res := <-term:
// 		cont = false
// 		return res
// 	case <-time.After(time.Duration(timeout) * time.Second):
// 		return nil
// 	}
// }
