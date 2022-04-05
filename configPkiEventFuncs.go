package shoset

import (
	// "fmt"
	// "crypto/tls"
	// "crypto/x509"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	// "time"

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

	if c.ch.GetIsPki() {
		// si je suis pki :
		//   si on m'envoie un certreq
		//   alors
		//     on extrait le certreq et le secret
		//     je signe
		//     je renvoie le resultat en precisant adresse dans un champ
		//     je reprend l'uuid du msg, je lui ajoute un caractere au bout (uuid_response) et je l'utilise comme uuid du msg de reponse
		//     return
		//   fi
		if evt.GetCertReq() != nil {
			/////////////////////////////
			CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.crt")
			if err != nil {
				return err
			}
			signedCert := c.ch.SignCertificate(evt.GetCertReq(), evt.GetHostPublicKey())
			if signedCert != nil {
				var returnPkiEvent *msg.PkiEvent

				
				////////////////////////////////

				if c.ch.GetLogicalName() == evt.GetLogicalName() { // les clusters deviennent à leur tour pki
					CAprivateKeyBytes, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/privateCAKey.key")
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
					returnPkiEvent = msg.NewPkiEventReturn(evt.GetRequestAddress(), signedCert, CAcert, CAprivateKey)
				} else {
					returnPkiEvent = msg.NewPkiEventReturn(evt.GetRequestAddress(), signedCert, CAcert, nil)
				}
				returnPkiEvent.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events
				SendPkiEvent(c.ch, returnPkiEvent)
			}
		}
	} else if c.ch.GetBindAddress() == evt.GetRequestAddress() {
		// si le msg est une reponse à ma demande (champ adresse equivaut la mienne), c'est donc moi qui ai envoyé le certreq
		// alors
		//   je recupere le msg et lire mon cert
		//   return
		// fi

		///////////////////////////////
		// fmt.Println(c.ch.ConnsByName.GetConfigName(), ":", evt.GetCAcert())

		if evt.GetSignedCert() != nil {
			signedCert := evt.GetSignedCert()
			certFile, err := os.Create(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/cert.crt")
			if err != nil {
				return err
			}
			pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCert})
			certFile.Close()
		}
		

		if evt.GetCAcert() != nil {
			caCert := evt.GetCAcert()
			ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/CAcert.crt", caCert, 0644)
		}

		///////////////////////////////
		
		if evt.GetCAprivateKey() != nil {
			caPrivateKey := evt.GetCAprivateKey()
			CAprivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey)})
			CAprivateKeyFile.Close()

			c.ch.SetIsPki(true)
		}
		c.ch.SetIsCertified(true)
	} else {
		// je transmet le msg puisque je suis ni pki ni demandeur
		if state := c.GetCh().Queue["pkievt"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
			SendPkiEvent(c.ch, evt)
		}
	}
	return nil
}

// SendEventConn :
func SendPkiEventConn(c *ShosetConn, evt interface{}) {
	c.WriteString("pkievt")
	c.WriteMessage(evt)
}

// SendEvent :
func SendPkiEvent(c *Shoset, evt msg.Message) {
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			conn.SendMessage(evt)
		},
	)
}
