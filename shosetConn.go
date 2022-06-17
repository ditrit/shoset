package shoset

import (
	"crypto/rsa"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ditrit/shoset/msg"
	uuid "github.com/kjk/betterguid"
	"github.com/rs/zerolog"
)

// ShosetConn : client connection
type ShosetConn struct {
	logger           zerolog.Logger
	socket           *tls.Conn
	remoteLname      string // logical name of the socket in front of this one
	remoteShosetType string // shosetType of the socket in fornt of this one
	dir              string
	remoteAddress    string // addresse of the socket in fornt of this one
	ch               *Shoset
	rb               *msg.Reader
	wb               *msg.Writer
	isValid          bool // for join protocol
}

func (c *ShosetConn) GetDir() string               { return c.dir }
func (c *ShosetConn) GetCh() *Shoset               { return c.ch }
func (c *ShosetConn) GetLocalLogicalName() string  { return c.ch.GetLogicalName() }
func (c *ShosetConn) GetRemoteLogicalName() string { return c.remoteLname }
func (c *ShosetConn) GetLocalShosetType() string   { return c.ch.GetShosetType() }
func (c *ShosetConn) GetRemoteShosetType() string  { return c.remoteShosetType }
func (c *ShosetConn) GetLocalAddress() string      { return c.ch.GetBindAddress() }
func (c *ShosetConn) GetRemoteAddress() string     { return c.remoteAddress }
func (c *ShosetConn) GetIsValid() bool             { return c.isValid }

func (c *ShosetConn) SetRemoteLogicalName(lName string)     { c.remoteLname = lName }
func (c *ShosetConn) SetLocalAddress(bindAddress string)    { c.ch.SetBindAddress(bindAddress) }
func (c *ShosetConn) SetRemoteShosetType(ShosetType string) { c.remoteShosetType = ShosetType }
func (c *ShosetConn) SetIsValid(state bool)                 { c.isValid = state }
func (c *ShosetConn) SetRemoteAddress(address string)       { c.remoteAddress = address }

func NewShosetConn(c *Shoset, address, dir string) (*ShosetConn, error) {
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}

	lgr := c.logger.With().Str("scid", uuid.New()).Logger()
	lgr.Debug().Strs("address-dir", []string{address, dir}).Msg("shosetConn created")
	return &ShosetConn{
		logger:        lgr,
		ch:            c,
		dir:           dir,
		socket:        new(tls.Conn),
		rb:            new(msg.Reader),
		wb:            new(msg.Writer),
		remoteAddress: ipAddress,
		isValid:       true,
	}, nil
}

func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{ localAddress: %s, name: %s, type: %s, way: %s, remoteAddress: %s}", c.GetLocalAddress(), c.GetRemoteLogicalName(), c.GetRemoteShosetType(), c.GetDir(), c.GetRemoteAddress())
}

func (c *ShosetConn) UpdateConn(conn *tls.Conn) {
	c.socket = conn
	c.rb.UpdateReader(conn)
	c.wb.UpdateWriter(conn)
}

// ReadString :
func (c *ShosetConn) ReadString() (string, error) {
	return c.rb.ReadString()
}

// WriteString :
func (c *ShosetConn) WriteString(data string) (int, error) {
	return c.wb.WriteString(data)
}

// ReadMessage :
func (c *ShosetConn) ReadMessage(data interface{}) error {
	return c.rb.ReadMessage(data)
}

// WriteMessage :
func (c *ShosetConn) WriteMessage(data interface{}) error {
	return c.wb.WriteMessage(data)
}

// Flush :
func (c *ShosetConn) Flush() error {
	return c.wb.Flush()
}

func (c *ShosetConn) runPkiRequest(address string) error {
	certReq, hostPublicKey, hostPrivateKey, err := PrepareCertificate()
	if err != nil {
		c.logger.Error().Msg("prepare certificate didn't work : " + err.Error())
		return errors.New("prepare certificate didn't work" + err.Error())
	}

	// Private key
	err = EncodeFile(hostPrivateKey, RSA_PRIVATE_KEY, c.ch.config.GetBaseDir()+c.ch.config.GetFileName()+PATH_PRIVATE_KEY)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return err
	}

	PkiEvent := msg.NewPkiEventInit(TLS_SINGLE_WAY_PKI_EVT, address, c.ch.GetLogicalName(), certReq, hostPublicKey)
	for {
		conn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.ch.tlsConfigSingleWay)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		c.UpdateConn(conn)

		err = c.SendMessage(*PkiEvent)
		if err != nil {
			c.logger.Error().Msg("couldn't send pkievt_TLSsingleWay: " + err.Error())
		}

		// receive messages
		err = c.receiveMessage()
		time.Sleep(time.Millisecond * time.Duration(100))
		if err != nil {
			continue
		}

		c.logger.Debug().Msg("runPkiRequest: socket closed")
		c.socket.Close()
		return nil
	}
}

// RunJoinConn : handler for the socket, for Join()
func (c *ShosetConn) runJoinConn() {
	joinConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, PROTOCOL_JOIN) //we create a new message config
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.ch.tlsConfigDoubleWay)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			c.logger.Error().Msg("join err: " + err.Error())
			continue
		}
		defer func() {
			c.logger.Debug().Msg("runJoinConn: socket closed")
			c.socket.Close()
		}()
		// a connection occured
		c.UpdateConn(conn)

		err = c.SendMessage(*joinConfig)
		if err != nil {
			c.logger.Error().Msg("couldn't send joincfg: " + err.Error())
			continue
		}

		// receive messages
		for {
			err := c.receiveMessage()
			time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.logger.Error().Msg("socket closed: err in receiveMessage join: " + err.Error())
				c.SetRemoteLogicalName(VOID) // reinitialize conn
				break
			}
		}

	}
}

// RunOutConn : handler for the socket, for Link()
func (c *ShosetConn) runLinkConn() {
	linkConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, PROTOCOL_LINK)
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.ch.tlsConfigDoubleWay)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		defer func() {
			c.logger.Debug().Msg("runLinkConn: socket closed")
			c.socket.Close()
		}()
		c.UpdateConn(conn)

		err = c.SendMessage(*linkConfig)
		if err != nil {
			c.logger.Error().Msg("couldn't send linkcfg: " + err.Error())
			continue
		}

		// receive messages
		for {
			err := c.receiveMessage()
			time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.logger.Error().Msg("socket closed: err in receiveMessage link: " + err.Error())
				c.SetRemoteLogicalName(VOID) // reinitialize conn
				break
			}
		}

	}
}

// runByeConn : handler for the socket, for Bye()
func (c *ShosetConn) runByeConn() {
	byeConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, PROTOCOL_EXIT) //we create a new message config
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.ch.tlsConfigDoubleWay)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		defer func() {
			c.logger.Debug().Msg("runByeConn: socket closed")
			c.socket.Close()
		}()
		// a connection occured
		c.UpdateConn(conn)

		err = c.SendMessage(*byeConfig)
		if err != nil {
			c.logger.Error().Msg("couldn't send byecfg: " + err.Error())
			continue
		}

		// receive messages
		for {
			err := c.receiveMessage()
			time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.logger.Error().Msg("socket closed: err in receiveMessage bye: " + err.Error())
				c.SetRemoteLogicalName(VOID) // reinitialize conn
				break
			}
		}

	}
}

func (c *ShosetConn) runInConnSingle(address string) {
	c.ch.ConnsSingleBool.Delete(address)

	// receive messages
	err := c.receiveMessage()
	time.Sleep(time.Millisecond * time.Duration(100))
	if err != nil {
		c.logger.Error().Msg("socket closed: err in receiveMessage inconnsingle: " + err.Error())
		c.SetRemoteLogicalName(VOID) // reinitialize conn
		return
	}
}

// runInConnDouble : handler for the connection, for handleBind()
func (c *ShosetConn) runInConnDouble() {
	defer func() {
		c.logger.Debug().Msg("double_way: socket closed")
		c.socket.Close()
	}()

	// receive messages
	for {
		err := c.receiveMessage()
		// time.Sleep(time.Millisecond * time.Duration(10))
		runtime.Gosched()
		if err != nil {
			c.logger.Error().Msg("err in receivmsg runInConnDouble: " + err.Error())
			if err.Error() == "Invalid connection for join - not the same type/name or shosetConn ended" {
				c.ch.SetIsValid(false)
			}
			return
		}
	}
}

// SendMessage :
func (c *ShosetConn) SendMessage(msg msg.Message) error {
	_, err := c.WriteString(msg.GetMsgType())
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			// https://gosamples.dev/broken-pipe/
			return nil
		} else if errors.Is(err, syscall.ECONNRESET) {
			// https://gosamples.dev/connection-reset-by-peer/
			return nil
		}
		return err
	}
	err = c.WriteMessage(msg)
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			return nil
		} else if errors.Is(err, syscall.ECONNRESET) {
			return nil
		}
		return err
	}
	return nil
}

func (c *ShosetConn) receiveMessage() error {
	if !c.GetIsValid() {
		c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		return errors.New("error : Invalid connection for join - not the same type/name or shosetConn ended")
	}

	// read message type
	msgType, err := c.rb.ReadString()
	//fmt.Println("msgType (receiveMessage) : ", msgType) //
	switch {
	case err == io.EOF:
		if c.GetDir() == IN {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return nil
	case errors.Is(err, syscall.ECONNRESET):
		return nil
	case errors.Is(err, syscall.EPIPE):
		return nil
	case err != nil:
		if c.GetDir() == IN {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New(err.Error())
	}
	msgType = strings.Trim(msgType, "\n")
	runtime.Gosched()

	if msgType == TLS_DOUBLE_WAY_TEST_WRITE { // do not handle this message because it's just a test for handleBind()
		return nil
	}

	err = c.handleMsg(msgType)
	if err != nil {
		return err
	}
	return nil
}

func (c *ShosetConn) handleMsg(msgType string) error {
	// read Message Value
	handler, ok := c.ch.Handlers[msgType]
	if !ok {
		if c.GetDir() == IN {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("receiveMessage : non implemented type of message " + msgType)
	}
	msgVal, err := handler.Get(c)
	//fmt.Println("msgVal (handleMsg) : ", msgVal) //
	if err != nil {
		if c.GetDir() == IN {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("receiveMessage : can not read value of " + msgType + " : " + err.Error())
	}
	//fmt.Println("msgType (handleMsg) : ", msgType) //

	DoubleWayMsgTypes := []string{"cfgjoin", "cfglink", "cfgbye", "pkievt_TLSdoubleWay","evt"}
	switch {
	case msgType == TLS_SINGLE_WAY_PKI_EVT:
		err := c.handleSingleWay(msgVal)
		if err != nil {
			return err
		}
	case contains(DoubleWayMsgTypes, msgType):
		err := handler.HandleDoubleWay(c, msgVal)
		if err != nil {
			return err
		}
	default:
		c.logger.Error().Msg("wrong msgType : " + msgType)
		return errors.New("wrong msgType : " + msgType)
	}
	return nil
}

func (c *ShosetConn) handleSingleWay(msgVal msg.Message) error {
	cmd := msgVal.GetMsgType()
	evt := msgVal.(msg.PkiEvent)

	switch {
	case evt.GetCommand() == TLS_SINGLE_WAY_PKI_EVT:
		err := c.handlePkiRequest(msgVal)
		if err != nil {
			return err
		}
	case evt.GetCommand() == TLS_SINGLE_WAY_RETURN_PKI_EVT:
		err := c.handlePkiResponse(msgVal)
		if err != nil {
			return err
		}
	default:
		c.logger.Error().Msg("wrong cmd : " + cmd)
		return errors.New("wrong cmd : " + cmd)
	}
	return nil
}

func (c *ShosetConn) handlePkiRequest(msgVal msg.Message) error {
	evt := msgVal.(msg.PkiEvent)
	// 2. un nouveau se connecte à moi et je suis passe plat
	if !c.ch.GetIsPki() {
		c.ch.ConnsSingleConn.Store(evt.GetRequestAddress(), c)
		handler := c.ch.Handlers["pkievt_TLSsingleWay"]
		evt.SetCommand(TLS_DOUBLE_WAY_PKI_EVT)
		handler.Send(c.ch, evt)
		return nil
	}

	// 1. un nouveau se connecte directement à moi et je suis PKI
	certDir := c.ch.config.GetBaseDir() + c.ch.config.GetFileName()
	CAcert, err := ioutil.ReadFile(certDir + PATH_CA_CERT)
	if err != nil {
		c.logger.Error().Msg("couldn't get CAcert: " + err.Error())
		return err
	}

	signedCert := c.ch.SignCertificate(evt.GetCertReq(), evt.GetHostPublicKey())
	if signedCert == nil {
		c.logger.Error().Msg("signCertificate didn't work")
		return errors.New("signCertificate didn't work")
	}

	var CAprivateKey *rsa.PrivateKey
	if c.ch.GetLogicalName() == evt.GetLogicalName() { // les clusters deviennent à leur tour pki
		CAprivateKey, err = GetPrivateKey(c.ch.config.GetBaseDir() + c.ch.config.GetFileName() + PATH_CA_PRIVATE_KEY)
		if err != nil {
			c.logger.Error().Msg("couldn't get CAprivateKey : " + err.Error())
			return err
		}
	}

	returnPkiEvent := msg.NewPkiEventReturn(TLS_SINGLE_WAY_RETURN_PKI_EVT, evt.GetRequestAddress(), signedCert, CAcert, CAprivateKey)
	returnPkiEvent.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events

	err = c.SendMessage(*returnPkiEvent)
	if err != nil {
		c.logger.Error().Msg("couldn't send singleConn returnpkievt: " + err.Error())
		return err
	}

	// 3. j'ai reçu un message autre que pkievt_TLSsingleWay, donc j'ignore
	return nil
}

func (c *ShosetConn) handlePkiResponse(msgVal msg.Message) error {
	evt := msgVal.(msg.PkiEvent)
	err := EncodeFile(evt.GetSignedCert(), CERTIFICATE, c.ch.config.GetBaseDir()+c.ch.config.GetFileName()+PATH_CERT)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return err
	}

	CApath := c.ch.config.GetBaseDir() + c.ch.config.GetFileName() + PATH_CA_CERT
	err = ioutil.WriteFile(CApath, evt.GetCAcert(), 0644)
	if err != nil {
		c.logger.Error().Msg("couldn't write CAcert: " + err.Error())
		return err
	}

	if evt.GetCAprivateKey() != nil {
		err = EncodeFile(evt.GetCAprivateKey(), RSA_PRIVATE_KEY, c.ch.config.GetBaseDir()+c.ch.config.GetFileName()+PATH_CA_PRIVATE_KEY)
		if err != nil {
			c.logger.Error().Msg(err.Error())
			return err
		}
		c.ch.SetIsPki(true)
	}

	// point env variable to our CAcert so that computer does not point elsewhere
	os.Setenv(CERT_FILE_ENV, c.ch.config.GetBaseDir()+c.ch.config.GetFileName()+PATH_CA_CERT)
	// tls Double way
	err = c.ch.SetUpDoubleWay(c.ch.config.GetBaseDir(), c.ch.config.GetFileName(), CApath)
	if err != nil {
		c.logger.Error().Msg(err.Error())
		return err
	}

	return nil
}
