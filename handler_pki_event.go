package shoset

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

type PkiEventHandler struct{}

// GetConfigPkiEvent :
func (ceh *PkiEventHandler) Get(c *ShosetConn) (msg.Message, error) {
	var evt msg.PkiEvent
	err := c.ReadMessage(&evt)
	return evt, err
}

// HandleConfigPkiEvent :
func (ceh *PkiEventHandler) Handle(c *ShosetConn, message msg.Message) error {
	evt := message.(msg.PkiEvent)
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}

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
							c.ch.logger.Error().Msg("err in decrypt : " + err.Error())
							return err
						}
					}
					CAprivateKey, err := x509.ParsePKCS1PrivateKey(b)
					if err != nil {
						c.ch.logger.Error().Msg("err in parse private key : " + err.Error())
						return err
					}
					returnPkiEvent = msg.NewPkiEventReturn("return_pkievt", evt.GetRequestAddress(), signedCert, CAcert, CAprivateKey)
				} else {
					returnPkiEvent = msg.NewPkiEventReturn("return_pkievt", evt.GetRequestAddress(), signedCert, CAcert, nil)
				}
				returnPkiEvent.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events
				ceh.Send(c.ch, returnPkiEvent)

			}
		}
	} else if conn := c.ch.ConnsSingleAddress.Get(evt.GetRequestAddress()); conn != nil && evt.GetCommand() == "return_pkievt" {
		// si le msg est une reponse à ma demande (champ adresse equivaut la mienne), c'est donc moi qui ai envoyé le certreq
		// alors
		//   je recupere le msg et lire mon cert
		//   return
		// fi

		// 5. je reçois une signedReq et je suis relié en singleWay au demandeur

		// c.SendMessage(evt)
		err := conn.SendMessage(evt)
		if err != nil {
			conn.ch.logger.Warn().Msg("couldn't send returnpkievt : " + err.Error())
			return err
		}
		// c.ch.ConnsSingleAddress[evt.GetRequestAddress()].socket.Close()
		c.socket.Close()
		c.ch.ConnsSingleAddress.Delete(evt.GetRequestAddress())
	} else {
		// je transmet le msg puisque je suis ni pki ni demandeur
		// 6. une shoset en double way me transmet une reqcert et je suis passe plat
		if state := c.GetCh().Queue["pkievt"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
			ceh.Send(c.ch, evt)
		}
	}
	return nil
}

// SendConfigPkiEvent :
func (ceh *PkiEventHandler) Send(c *Shoset, evt msg.Message) {
	c.ConnsByName.IterateAll(
		func(key string, conn *ShosetConn) {
			// conn.rb = msg.NewReader(conn.socket)
			// conn.wb = msg.NewWriter(conn.socket)
			if err := conn.SendMessage(evt); err != nil {
				conn.ch.logger.Warn().Msg("couldn't send pkievt : " + err.Error())
			}
		},
	)
}

// WaitConfigPkiEvent :
func (ceh *PkiEventHandler) Wait(c *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("PkiEventHandler.Wait not implemented")
	return nil
}
