package shoset

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/ditrit/shoset/msg"
	"github.com/rs/zerolog/log"
)

// PkiEventHandler implements MessageHandlers interface.
type PkiEventHandler struct{}

// Get returns the message for a given ShosetConn.
func (peh *PkiEventHandler) Get(c *ShosetConn) (msg.Message, error) {
	var evt msg.PkiEvent
	err := c.GetReader().ReadMessage(&evt)
	return evt, err
}

// RunPkiRequest is the first step to get certified.
// Creates certificateRequest and send it as an event in the network to reach CA and get certified.
func (c *ShosetConn) RunPkiRequest(address string) error {
	certificateRequest, hostPublicKey, hostPrivateKey, err := PrepareCertificate()
	if err != nil {
		c.Logger.Error().Msg("prepare certificate didn't work : " + err.Error())
		return errors.New("prepare certificate didn't work" + err.Error())
	}

	err = EncodeFile(hostPrivateKey, RSA_PRIVATE_KEY, c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory()+c.GetShoset().ConnsByLname.GetConfig().GetFileName()+PATH_PRIVATE_KEY)
	if err != nil {
		c.Logger.Error().Msg(err.Error())
		return err
	}

	pkiEvent := msg.NewPkiEventInit(TLS_SINGLE_WAY_PKI_EVT, address, c.GetShoset().GetLogicalName(), certificateRequest, hostPublicKey)
	for {
		pkiConn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.GetShoset().tlsConfigSingleWay)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(10))
			continue
		}
		c.UpdateConn(pkiConn)

		err = c.GetWriter().SendMessage(*pkiEvent)
		if err != nil {
			c.Logger.Error().Msg("couldn't send pkievt_TLSsingleWay: " + err.Error())
		}

		err = c.ReceiveMessage()
		if err != nil {
			continue
		}

		c.Logger.Debug().Msg("RunPkiRequest: socket closed")
		c.GetConn().Close()
		return nil
	}
}

// HandleSingleWay handle receive message as a pki event.
// Then, based on the message's command, it handles message value adequately.
func (c *ShosetConn) HandleSingleWay(messageValue msg.Message) error {
	evt := messageValue.(msg.PkiEvent)
	switch {
	case evt.GetCommand() == TLS_SINGLE_WAY_PKI_EVT:
		err := c.HandlePkiRequest(messageValue)
		if err != nil {
			return err
		}
	case evt.GetCommand() == TLS_SINGLE_WAY_RETURN_PKI_EVT:
		err := c.HandlePkiResponse(messageValue)
		if err != nil {
			return err
		}
	default:
		c.Logger.Error().Msg("wrong cmd : " + evt.GetCommand())
		return errors.New("wrong cmd : " + evt.GetCommand())
	}
	return nil
}

// HandlePkiRequest handles a pki request.
// Depending if the Shoset is PKI or not, certify or send event back in network.
func (c *ShosetConn) HandlePkiRequest(messageValue msg.Message) error {
	evt := messageValue.(msg.PkiEvent)

	if !c.GetShoset().GetIsPki() {
		c.GetShoset().ConnsSingleConn.Store(evt.GetRequestAddress(), c)
		handler := c.GetShoset().Handlers[TLS_SINGLE_WAY_PKI_EVT]
		evt.SetCommand(TLS_DOUBLE_WAY_PKI_EVT)
		handler.Send(c.GetShoset(), evt)
		return nil
	}

	certificateDirectory := c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory() + c.GetShoset().ConnsByLname.GetConfig().GetFileName()
	CAcertificate, err := ioutil.ReadFile(certificateDirectory + PATH_CA_CERT)
	if err != nil {
		c.Logger.Error().Msg("couldn't get CAcertificate: " + err.Error())
		return err
	}

	signedCertificate := c.GetShoset().SignCertificate(evt.GetCertificateRequest(), evt.GetHostPublicKey())
	if signedCertificate == nil {
		c.Logger.Error().Msg("signCertificate didn't work")
		return errors.New("signCertificate didn't work")
	}

	var CAprivateKey *rsa.PrivateKey
	if c.GetShoset().GetLogicalName() == evt.GetLogicalName() { // same logical name than CA, add CAprivateKey to the return event
		CAprivateKey, err = GetPrivateKey(c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory() + c.GetShoset().ConnsByLname.GetConfig().GetFileName() + PATH_CA_PRIVATE_KEY)
		if err != nil {
			c.Logger.Error().Msg("couldn't get CAprivateKey : " + err.Error())
			return err
		}
	}

	pkiEventResponse := msg.NewPkiEventReturn(TLS_SINGLE_WAY_RETURN_PKI_EVT, evt.GetRequestAddress(), signedCertificate, CAcertificate, CAprivateKey)
	err = c.GetWriter().SendMessage(*pkiEventResponse)
	if err != nil {
		c.Logger.Error().Msg("couldn't send singleConn returnpkievt: " + err.Error())
		return err
	}
	return nil
}

// HandlePkiResponse encodes received signed certificate and sets up TLS Double Way for the Shoset.
func (c *ShosetConn) HandlePkiResponse(messageValue msg.Message) error {
	evt := messageValue.(msg.PkiEvent)
	err := EncodeFile(evt.GetSignedCert(), CERTIFICATE, c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory()+c.GetShoset().ConnsByLname.GetConfig().GetFileName()+PATH_CERT)
	if err != nil {
		c.Logger.Error().Msg(err.Error())
		return err
	}

	CApath := c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory() + c.GetShoset().ConnsByLname.GetConfig().GetFileName() + PATH_CA_CERT
	err = ioutil.WriteFile(CApath, evt.GetCAcert(), 0644)
	if err != nil {
		c.Logger.Error().Msg("couldn't write CAcertificate: " + err.Error())
		return err
	}

	if evt.GetCAprivateKey() != nil { // same logical name than CA, so the Shoset becomes CA as well
		err = EncodeFile(evt.GetCAprivateKey(), RSA_PRIVATE_KEY, c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory()+c.GetShoset().ConnsByLname.GetConfig().GetFileName()+PATH_CA_PRIVATE_KEY)
		if err != nil {
			c.Logger.Error().Msg(err.Error())
			return err
		}
		c.GetShoset().SetIsPki(true)
	}

	os.Setenv(CERT_FILE_ENVIRONMENT, c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory()+c.GetShoset().ConnsByLname.GetConfig().GetFileName()+PATH_CA_CERT)
	err = c.GetShoset().SetUpDoubleWay()
	if err != nil {
		c.Logger.Error().Msg(err.Error())
		return err
	}
	return nil
}

// HandleDoubleWay handles message for a ShosetConn accordingly.
func (peh *PkiEventHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	evt := message.(msg.PkiEvent)

	singleWayCertReqConn, _ := c.GetShoset().ConnsSingleConn.Load(evt.GetRequestAddress())
	switch {
	case evt.GetCommand() == TLS_DOUBLE_WAY_PKI_EVT && c.GetShoset().GetIsPki():
		// this shoset is PKI and a TLSdoubleWay shoset sent a cert request.
		// handle this cert request and return the signed cert.
		if evt.GetCertificateRequest() == nil {
			c.Logger.Error().Msg("empty cert req received")
			return errors.New("empty cert req received")
		}

		cfgDir := c.GetShoset().ConnsByLname.GetConfig().GetBaseDirectory()
		CAcertificate, err := ioutil.ReadFile(cfgDir + c.GetShoset().ConnsByLname.GetConfig().GetFileName() + PATH_CA_CERT)
		if err != nil {
			return err
		}

		signedCertificate := c.GetShoset().SignCertificate(evt.GetCertificateRequest(), evt.GetHostPublicKey())
		if signedCertificate == nil {
			c.Logger.Error().Msg("CA failed to sign cert")
			return errors.New("CA failed to sign cert")
		}

		pkiEventResponse := msg.NewPkiEventReturn("return_pkievt_TLSdoubleWay", evt.GetRequestAddress(), signedCertificate, CAcertificate, nil)
		pkiEventResponse.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events.
		peh.Send(c.GetShoset(), *pkiEventResponse)

	case evt.GetCommand() == "return_pkievt_TLSdoubleWay" && singleWayCertReqConn != nil:
		// this shoset received a signed cert from a TLSdoubleWay shoset destined to a TLSsingleWay shoset known by this one.
		// send back the signed cert to the destined TLSsingleWay shoset.
		evt.SetCommand(TLS_SINGLE_WAY_RETURN_PKI_EVT)
		if err := singleWayCertReqConn.(*ShosetConn).GetWriter().SendMessage(evt); err != nil {
			singleWayCertReqConn.(*ShosetConn).Logger.Warn().Msg("couldn't send returnpkievt : " + err.Error())
			return err
		}
		c.GetShoset().ConnsSingleConn.Delete(evt.GetRequestAddress())

	default:
		// this shoset is not PKI and a TLSdoubleWay shoset sent a cert request.
		// sends back the cert request into the network until it reaches a PKI if it is not in it yet.
		if notInQueue := c.GetShoset().Queue["pkievt_TLSdoubleWay"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); notInQueue {
			peh.Send(c.GetShoset(), evt)
		}
	}
	return nil
}

// Send sends the message through the given Shoset network.
func (peh *PkiEventHandler) Send(s *Shoset, evt msg.Message) {
	s.ConnsByLname.Iterate(
		func(key string, conn interface{}) {
			if err := conn.(*ShosetConn).GetWriter().SendMessage(evt); err != nil {
				conn.(*ShosetConn).Logger.Warn().Msg("couldn't send pkievt_TLSsingleWay : " + err.Error())
			}
		},
	)
}

// Wait returns the message received for a given Shoset.
func (peh *PkiEventHandler) Wait(s *Shoset, replies *msg.Iterator, args map[string]string, timeout int) *msg.Message {
	// no-op
	log.Warn().Msg("PkiEventHandler.Wait not implemented")
	return nil
}
