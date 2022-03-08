package shoset

import (
	"fmt"
	"io/ioutil"
	"os"

	// "fmt"

	"github.com/ditrit/shoset/msg"
	"github.com/square/certstrap/pkix"
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
	//dir := c.GetDir()
	remoteAddress := cfg.GetAddress()

	switch cfg.GetCommandName() {
	case "pki":
		if c.ch.GetIsPki() {

			dirname, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Get UserHomeDir error : ", err)
			}
			cert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.pem")
			if err != nil {
				fmt.Println("error in reading file")
			}

			// fmt.Println(ch)

			// fmt.Println(ch.GetBindAddress(), "enters config pki from", remoteAddress)
			// fmt.Println(remoteAddress)
			configPki := msg.NewCfgPki(remoteAddress, ch.GetLogicalName(), ch.GetShosetType(), "return_pki", cert)
			c.SendMessage(configPki)
			// fmt.Println(ch.GetBindAddress(), "message sent to ", remoteAddress)
		} else if cfg.GetLogicalName() == "Ca" {
			fmt.Println("connector detected")
		} else {
			// tant que la chaussette en face n'est pas pki on demande une nouvelle config
			newPkiConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "pki")
			c.SendMessage(newPkiConfig)
		}
	case "return_pki":
		if !c.ch.GetIsCertified() {
			dirname, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Get UserHomeDir error : ", err)
			}

			CAcertFile, err := os.Create(dirname + "/.shoset/" + c.ch.ConnsByName.GetConfigName() + "/cert/CAcert.pem")
			if err != nil {
				fmt.Println("Create CA certificate file error : ", err)
			}

			_, err = CAcertFile.Write(caCert)
			if err != nil {
				fmt.Println("Write in CA certificate file error : ", err)
			}

			// génération des clefs privée, publique et request pour la shoset
			hostKey := c.ch.CreateKey()
			// création du certificat signé avec la clef privée de la CA
			hostCsr := c.ch.CreateSignRequest(hostKey)

			newCAcert, err := pkix.NewCertificateFromPEM(caCert) // https://github.com/square/certstrap/blob/v1.2.0/pkix/cert.go#L48
			if err != nil {
				fmt.Println("New CA certificate file error : ", err)
			}
			c.ch.SignRequest(newCAcert, hostCsr, hostKey)

			if c.ch.GetLogicalName() == cfg.GetLogicalName() {

				c.ch.SetIsPki(true)
				// fmt.Println(ch.GetBindAddress(), "is now pki")
				// return nil
			} else {
				fmt.Println(ch.GetBindAddress(), "case not treated yet")
			}
			c.ch.SetIsCertified(true)
		}
	}
	return nil
}
