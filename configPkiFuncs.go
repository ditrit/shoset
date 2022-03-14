package shoset

import (
	// "crypto/x509"
	// "encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ditrit/shoset/msg"
	// "github.com/square/certstrap/pkix"
)

// GetConfigJoin :
func GetConfigPki(c *ShosetConn) (msg.Message, error) {
	var cfg msg.ConfigProtocol
	err := c.ReadMessage(&cfg)
	return cfg, err
}

// HandleConfigJoin :
func HandleConfigPki(c *ShosetConn, message msg.Message) error {
	cfg := message.(msg.ConfigProtocol) // compute config from message
	ch := c.GetCh()
	caCert := cfg.GetCAcert()
	caPrivateKey := cfg.GetCAprivateKey()
	dir := c.GetDir()
	remoteAddress := cfg.GetAddress()

	switch cfg.GetCommandName() {
	case "pki":
		// fmt.Println(c.ch.GetBindAddress(), "enters pki for ", remoteAddress)
		// if dir == "in" {
		if c.ch.GetIsPki() && c.ch.GetLogicalName() == cfg.GetLogicalName() {
			dirname, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.crt")
			if err != nil {
				return err
			}

			// block, _ := pem.Decode([]byte(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.crt"))
			// CAcert, err := x509.ParseCertificate(block.Bytes)
			// if err != nil {
			// 	fmt.Println("couldn't decode cert :", err)
			// }

			CAprivateKey, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/privateCAKey.key")
			if err != nil {
				return err
			}

			// block2, _ := pem.Decode([]byte(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/privateCAKey.key"))
			// CAprivateKey, _ := x509.ParsePKCS1PrivateKey(block2.Bytes)

			// fmt.Println(ch)

			// fmt.Println(ch.GetBindAddress(), "enters config pki from", remoteAddress)
			// fmt.Println(remoteAddress)
			configPki := msg.NewCfgPki(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "return_pki", CAcert, CAprivateKey)
			c.SendMessage(configPki)
			// fmt.Println(ch.GetBindAddress(), "message sent to ", remoteAddress)
		} else {
			if c.ch.GetLogicalName() == cfg.GetLogicalName() {
				// tant que la chaussette en face n'est pas pki on demande une nouvelle config
				newPkiConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "pki")
				// fmt.Println(c.ch.GetBindAddress(), "creates cfg for ", remoteAddress)
				c.SendMessage(newPkiConfig)
			}

			// fmt.Println("create event")
			// pkiEvent := msg.NewEventClassic("pki", "pki_event", "ask_cert")
			// // SendEventConn(c, pkiEvent)
			// SendEvent(c.ch, pkiEvent)
			// fmt.Println("event sent")
		}

	case "return_pki":
		if dir == "out" {
			if !c.ch.GetIsCertified() {
				dirname, err := os.UserHomeDir()
				if err != nil {
					return err
				}

				// Public key
				// CAcertFile, err := os.Create(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.crt")
				// if err != nil {
				// 	return err
				// }
				ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/CAcert.crt", caCert, 0644)
				// block, _ := pem.Decode([]byte(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.crt"))
				// pem.Encode(CAcertFile, &pem.Block{Type: "CERTIFICATE", Bytes: block.Bytes})
				// CAcertFile.Close()
				// log.Print("written cert.pem\n")

				// Private key
				// CAprivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
				// if err != nil {
				// 	return err
				// }
				ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", caPrivateKey, 0644)
				// pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey)})
				// CAprivateKeyFile.Close()

				// // génération des clefs privée, publique et request pour la shoset
				// hostPublicKey, hostPrivateKey := c.ch.CreateKey()
				// // création du certificat signé avec la clef privée de la CA
				// hostCsr := c.ch.CreateSignRequest(hostKey)

				// newCAcert, err := pkix.NewCertificateFromPEM(caCert) // https://github.com/square/certstrap/blob/v1.2.0/pkix/cert.go#L48
				// if err != nil {
				// 	fmt.Println("New CA certificate file error : ", err)
				// }
				// c.ch.SignRequest(newCAcert, hostCsr, hostKey)

				fmt.Println(c.ch.GetBindAddress(), "enters initcertificate")
				err = c.ch.InitCertificate(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/CAcert.crt", dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key")
				if err != nil {
					fmt.Println("init certificate didn't work")
				}
				fmt.Println(c.ch.GetBindAddress(), "ended initcertificate")

				if c.ch.GetLogicalName() == cfg.GetLogicalName() {

					c.ch.SetIsPki(true)
					// fmt.Println(ch.GetBindAddress(), "is now pki")
					// return nil
				}
				c.ch.SetIsCertified(true)
			}
		}
	}
	return nil
}
