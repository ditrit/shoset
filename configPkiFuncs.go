package shoset

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ditrit/shoset/msg"
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

			CAprivateKey, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/privateCAKey.key")
			if err != nil {
				return err
			}

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
				ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/CAcert.crt", caCert, 0644)

				// Private key
				ioutil.WriteFile(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key", caPrivateKey, 0644)

				// fmt.Println(c.ch.GetBindAddress(), "enters initcertificate")
				err = c.ch.InitCertificate(dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/CAcert.crt", dirname+"/.shoset/"+c.ch.ConnsByName.GetConfigName()+"/cert/privateCAKey.key")
				if err != nil {
					fmt.Println("init certificate didn't work")
				}
				// fmt.Println(c.ch.GetBindAddress(), "ended initcertificate")

				if c.ch.GetLogicalName() == cfg.GetLogicalName() {

					c.ch.SetIsPki(true)
				}
				c.ch.SetIsCertified(true)
			}
		}
	}
	return nil
}
