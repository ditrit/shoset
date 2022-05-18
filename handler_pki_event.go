package shoset

import (
	"errors"
	"io/ioutil"

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
func (ceh *PkiEventHandler) HandleDoubleWay(c *ShosetConn, message msg.Message) error {
	evt := message.(msg.PkiEvent)

	conn, _ := c.ch.ConnsSingleConn.Load(evt.GetRequestAddress())
	switch {
	case evt.GetCommand() == TLS_DOUBLE_WAY_PKI_EVT && c.ch.GetIsPki():
		// this shoset is PKI and a TLSdoubleWay shoset sent a cert request
		// handle this cert request and return the signed cert
		if evt.GetCertReq() == nil {
			c.logger.Error().Msg("empty cert req received")
			return errors.New("empty cert req received")
		}

		cfgDir := c.ch.config.GetBaseDir()
		CAcert, err := ioutil.ReadFile(cfgDir + c.ch.config.GetFileName() + PATH_CA_CERT)
		if err != nil {
			return err
		}

		signedCert := c.ch.SignCertificate(evt.GetCertReq(), evt.GetHostPublicKey())
		if signedCert == nil {
			c.logger.Error().Msg("CA failed to sign cert")
			return errors.New("CA failed to sign cert")
		}

		returnPkiEvent := msg.NewPkiEventReturn("return_pkievt_TLSdoubleWay", evt.GetRequestAddress(), signedCert, CAcert, nil)
		returnPkiEvent.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events
		ceh.Send(c.ch, *returnPkiEvent)

	case evt.GetCommand() == "return_pkievt_TLSdoubleWay" && conn != nil:
		// this shoset received a signed cert from a TLSdoubleWay shoset destined to a TLSsingleWay shoset known by this one
		// send back the signed cert to the destined TLSsingleWay shoset
		evt.SetCommand(TLS_SINGLE_WAY_RETURN_PKI_EVT)
		if err := conn.(*ShosetConn).SendMessage(evt); err != nil {
			conn.(*ShosetConn).logger.Warn().Msg("couldn't send returnpkievt : " + err.Error())
			return err
		}
		c.ch.ConnsSingleConn.Delete(evt.GetRequestAddress())

	default:
		// this shoset is not PKI and a TLSdoubleWay shoset sent a cert request
		// send back the cert request into the network until it reaches a PKI
		if state := c.GetCh().Queue["pkievt_TLSdoubleWay"].Push(evt, c.GetRemoteShosetType(), c.GetLocalAddress()); state {
			ceh.Send(c.ch, evt)
		}
	}
	return nil
}

// SendConfigPkiEvent :
func (ceh *PkiEventHandler) Send(c *Shoset, evt msg.Message) {
	c.ConnsByName.Iterate(
		func(key string, conn interface{}) {
			if err := conn.(*ShosetConn).SendMessage(evt); err != nil {
				conn.(*ShosetConn).logger.Warn().Msg("couldn't send pkievt_TLSsingleWay : " + err.Error())
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
