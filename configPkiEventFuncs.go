package shoset

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
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

	// for k := range c.ch.ConnsSingleAddress {
	// 	fmt.Printf("key[%s] value[%s]\n", k, c.ch.ConnsSingleAddress[k])
	// 	fmt.Println(evt.GetRequestAddress())
	// }

	if c.ch.GetIsPki() && evt.GetCommand() == "pkievt" {
		// si je suis pki :
		//   si on m'envoie un certreq
		//   alors
		//     on extrait le certreq et le secret
		//     je signe
		//     je renvoie le resultat en precisant adresse dans un champ
		//     je reprend l'uuid du msg, je lui ajoute un caractere au bout (uuid_response) et je l'utilise comme uuid du msg de reponse
		//     return
		//   fi

		// 4. une shoset en double way me transmet une certreq et je suis pki
		// fmt.Println("received event 2")
		if evt.GetCertReq() != nil {
			CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/CAcert.crt")
			if err != nil {
				return err
			}
			signedCert := c.ch.SignCertificate(evt.GetCertReq(), evt.GetHostPublicKey())
			if signedCert != nil {
				var returnPkiEvent *msg.PkiEvent

				if c.ch.GetLogicalName() == evt.GetLogicalName() { // les clusters deviennent à leur tour pki
					CAprivateKeyBytes, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/privateCAKey.key")
					if err != nil {
						return err
					}
					block, _ := pem.Decode(CAprivateKeyBytes)
					enc := x509.IsEncryptedPEMBlock(block)
					b := block.Bytes
					if enc {
						b, err = x509.DecryptPEMBlock(block, nil)
						if err != nil {
							fmt.Println(err)
						}
					}
					CAprivateKey, err := x509.ParsePKCS1PrivateKey(b)
					if err != nil {
						fmt.Println(err)
					}
					returnPkiEvent = msg.NewPkiEventReturn("return_pkievt", evt.GetRequestAddress(), signedCert, CAcert, CAprivateKey)
				} else {
					returnPkiEvent = msg.NewPkiEventReturn("return_pkievt", evt.GetRequestAddress(), signedCert, CAcert, nil)
				}
				returnPkiEvent.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events
				// fmt.Println("return msg 2 sent to ", evt.GetRequestAddress())
				SendPkiEvent(c.ch, returnPkiEvent)

			}
		}
	} else if conn := c.ch.ConnsSingleAddress.Get(evt.GetRequestAddress()); conn != nil && evt.GetCommand() == "return_pkievt" {
		// si le msg est une reponse à ma demande (champ adresse equivaut la mienne), c'est donc moi qui ai envoyé le certreq
		// alors
		//   je recupere le msg et lire mon cert
		//   return
		// fi

		// 5. je reçois une signedReq et je suis relié en singleWay au demandeur

		// fmt.Println("I'm ", c.ch.GetBindAddress(), " and I have received a msg for", evt.GetRequestAddress())
		// c.SendMessage(evt)
		conn.SendMessage(evt)
		// c.ch.ConnsSingleAddress[evt.GetRequestAddress()].socket.Close()
		c.socket.Close()
		c.ch.ConnsSingleAddress.Delete(evt.GetRequestAddress())

		// fmt.Println("msg sent to ", c.ch.ConnsSingleAddress[evt.GetRequestAddress()].GetLocalAddress(), c.ch.ConnsSingleAddress[evt.GetRequestAddress()].GetRemoteAddress())

		// if evt.GetSignedCert() != nil && evt.GetCAcert() != nil {
		// 	signedCert := evt.GetSignedCert()
		// 	certFile, err := os.Create(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/cert.crt")
		// 	if err != nil {
		// 		return err
		// 	}
		// 	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCert})
		// 	certFile.Close()

		// 	caCert := evt.GetCAcert()
		// 	ioutil.WriteFile(dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/CAcert.crt", caCert, 0644)

		// 	if evt.GetCAprivateKey() != nil {
		// 		caPrivateKey := evt.GetCAprivateKey()
		// 		CAprivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey)})
		// 		CAprivateKeyFile.Close()

		// 		c.ch.SetIsPki(true)
		// 	}
		// 	c.ch.SetIsCertified(true)

		// 	// point env variable to our CAcert so that computer does not point elsewhere
		// 	os.Setenv("SSL_CERT_FILE", dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/CAcert.crt")

		// 	// tls Double way
		// 	cert, err := tls.LoadX509KeyPair(dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/cert.crt", dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/privateKey.key")
		// 	if err != nil {
		// 		fmt.Println("! Unable to Load certificate !")
		// 	}
		// 	CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/CAcert.crt")
		// 	if err != nil {
		// 		fmt.Println("error read file cacert :", err)
		// 	}
		// 	caCertPool := x509.NewCertPool()
		// 	caCertPool.AppendCertsFromPEM(CAcert)
		// 	c.ch.tlsConfigDoubleWay = &tls.Config{
		// 		Certificates:       []tls.Certificate{cert},
		// 		ClientCAs:          caCertPool,
		// 		ClientAuth:         tls.RequireAndVerifyClientCert,
		// 		InsecureSkipVerify: false,
		// 	}
		// 	c.ch.tlsConfigDoubleWay.BuildNameToCertificate()

		// 	// tls config single way
		// 	c.ch.tlsConfigSingleWay = &tls.Config{
		// 		Certificates:       []tls.Certificate{cert},
		// 		InsecureSkipVerify: false,
		// 	}
		// }
	} else {
		// je transmet le msg puisque je suis ni pki ni demandeur
		// 6. une shoset en double way me transmet une reqcert et je suis passe plat
		if state := c.GetCh().Queue["pkievt"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
			// fmt.Println(c.ch.GetPkiRequestAddress(), "transmits msg from ", evt.GetRequestAddress())
			SendPkiEvent(c.ch, evt)
		}
	}
	return nil
}

// SendEventConn :
func SendPkiEventConn(c *ShosetConn, evt interface{}) {
	_, err := c.WriteString("pkievt")
	if err != nil {
		fmt.Println("couldn't write string pkievt")
		return
	}
	err = c.WriteMessage(evt)
	if err != nil {
		fmt.Println("couldn't write pkievt")
		return
	}
}

// SendEvent :
func SendPkiEvent(c *Shoset, evt msg.Message) {
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			conn.SendMessage(evt)
		},
	)
}
