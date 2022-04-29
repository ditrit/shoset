package shoset

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	//	uuid "github.com/kjk/betterguid"

	"github.com/ditrit/shoset/msg"
)

// ShosetConn : client connection
type ShosetConn struct {
	socket           *tls.Conn
	remoteLname      string // logical name of the socket in front of this one
	remoteShosetType string // shosetType of the socket in fornt of this one
	dir              string
	remoteAddress    string // addresse of the socket in fornt of this one
	ch               *Shoset
	rb               *msg.Reader
	wb               *msg.Writer
	isValid          bool // for join protocol
	mu               sync.Mutex
}

// GetDir :
func (c *ShosetConn) GetDir() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dir
}

// GetCh :
func (c *ShosetConn) GetCh() *Shoset {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ch
}

func (c *ShosetConn) GetLocalLogicalName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ch.GetLogicalName()
}

// GetName : // remote logical Name
func (c *ShosetConn) GetRemoteLogicalName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.remoteLname
}

func (c *ShosetConn) GetLocalShosetType() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ch.GetShosetType()
}

// GetShosetType : // remote ShosetTypeName
func (c *ShosetConn) GetRemoteShosetType() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.remoteShosetType
}

// GetBindAddr : port sur lequel on est bindé
func (c *ShosetConn) GetLocalAddress() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ch.GetBindAddress()
}

func (c *ShosetConn) GetRemoteAddress() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.remoteAddress
}

func (c *ShosetConn) GetIsValid() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isValid
}

// SetName : // remote logical Name
func (c *ShosetConn) SetRemoteLogicalName(lName string) { // remote logical Name
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remoteLname = lName // remote logical Name
}

// SetBindAddr :
func (c *ShosetConn) SetLocalAddress(bindAddress string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if bindAddress != "" {
		c.ch.SetBindAddress(bindAddress)
	}
}

// SetShosetType : // remote ShosetType
func (c *ShosetConn) SetRemoteShosetType(ShosetType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ShosetType != "" {
		c.remoteShosetType = ShosetType
	}
}

func (c *ShosetConn) SetIsValid(state bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isValid = state
}

func (c *ShosetConn) SetRemoteAddress(address string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if address != "" {
		c.remoteAddress = address
	}
}

func NewShosetConn(c *Shoset, address string, dir string) (*ShosetConn, error) {
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}

	return &ShosetConn{
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
	return fmt.Sprintf("ShosetConn{ localAddress : %s, name : %s, type : %s, way : %s, remoteAddress : %s}", c.GetLocalAddress(), c.GetRemoteLogicalName(), c.GetRemoteShosetType(), c.GetDir(), c.GetRemoteAddress())
}

// ReadString :
func (c *ShosetConn) ReadString() (string, error) {
	return c.rb.ReadString()
}

// ReadMessage :
func (c *ShosetConn) ReadMessage(data interface{}) error {
	return c.rb.ReadMessage(data)
}

// WriteString :
func (c *ShosetConn) WriteString(data string) (int, error) {
	return c.wb.WriteString(data)
}

// Flush :
func (c *ShosetConn) Flush() error {
	return c.wb.Flush()
}

// WriteMessage :
func (c *ShosetConn) WriteMessage(data interface{}) error {
	return c.wb.WriteMessage(data)
}

func (c *ShosetConn) runPkiRequest() error {
	certReq, hostPublicKey, _ := c.ch.PrepareCertificate()
	if certReq != nil && hostPublicKey != nil {
		PkiEvent := msg.NewPkiEventInit("pkievt", c.ch.GetPkiRequestAddress(), c.ch.GetLogicalName(), certReq, hostPublicKey)
		for {
			if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
				return errors.New("shoset not valid")
			}

			c.ch.logger.Info().Msg("gonna dial")
			conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfigSingleWay)
			c.ch.logger.Info().Msg("dial done")
			if err != nil {
				time.Sleep(time.Millisecond * time.Duration(100))
				c.ch.logger.Error().Msg(err.Error())
				continue
			}
			c.socket = conn
			c.rb.UpdateReader(c.socket)
			c.wb.UpdateWriter(c.socket)
			// defer conn.Close()

			err = c.SendMessage(*PkiEvent)
			if err != nil {
				c.ch.logger.Error().Msg("couldn't send pkievt : " + err.Error())
			}

			// receive messages
			for {
				msgType, err := c.rb.ReadString()
				if err != nil {
					// fmt.Println("# error : ", err)
					break
				}
				msgType = strings.Trim(msgType, "\n")
				runtime.Gosched()

				fGet, ok := c.ch.Get[msgType]
				if !ok {
					c.ch.logger.Error().Msg("not ok client")
					break
				}
				msgVal, err := fGet(c)
				if err != nil {
					c.ch.logger.Error().Msg("didn't find function to handle event")
					break
				}

				if cmd := msgVal.GetMsgType(); cmd != "pkievt" {
					c.ch.logger.Error().Msg("not the right command client : " + cmd)
					// c.socket.Close() //////////////////////////////////////////////////////////
					break
				}

				evt := msgVal.(msg.PkiEvent)
				if c.ch.GetPkiRequestAddress() != evt.GetRequestAddress() || evt.GetCommand() != "return_pkievt" {
					c.ch.logger.Error().Msg("return msg does not correspond")
					break
				}

				if evt.GetSignedCert() != nil && evt.GetCAcert() != nil {
					dirname, err := os.UserHomeDir()
					if err != nil {
						c.ch.logger.Error().Msg("did not find home dir : " + err.Error())
						return err
					}

					signedCert := evt.GetSignedCert()
					certFile, err := os.Create(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/cert.crt")
					if err != nil {
						c.ch.logger.Error().Msg("couldn't create certfile : " + err.Error())
						break
					}
					pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: signedCert})
					certFile.Close()

					caCert := evt.GetCAcert()
					err = ioutil.WriteFile(dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/CAcert.crt", caCert, 0644)
					if err != nil {
						c.ch.logger.Error().Msg("couldn't write CAcert : " + err.Error())
						break
					}

					if evt.GetCAprivateKey() != nil {
						caPrivateKey := evt.GetCAprivateKey()
						CAprivateKeyFile, err := os.OpenFile(dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/privateCAKey.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
						if err != nil {
							c.ch.logger.Error().Msg("couldn't create CAprivateKeyFile : " + err.Error())
							break
						}
						pem.Encode(CAprivateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey)})
						CAprivateKeyFile.Close()

						c.ch.SetIsPki(true)
					}
					c.ch.SetIsCertified(true)

					// point env variable to our CAcert so that computer does not point elsewhere
					os.Setenv("SSL_CERT_FILE", dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/CAcert.crt")

					// tls Double way
					cert, err := tls.LoadX509KeyPair(dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/cert.crt", dirname+"/.shoset/"+c.ch.GetFileName()+"/cert/privateKey.key")
					if err != nil {
						c.ch.logger.Error().Msg("Unable to Load certificate : " + err.Error())
					}
					CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/CAcert.crt")
					if err != nil {
						c.ch.logger.Error().Msg("error read file cacert : " + err.Error())
					}
					caCertPool := x509.NewCertPool()
					caCertPool.AppendCertsFromPEM(CAcert)
					c.ch.tlsConfigDoubleWay = &tls.Config{
						Certificates:       []tls.Certificate{cert},
						ClientCAs:          caCertPool,
						ClientAuth:         tls.RequireAndVerifyClientCert,
						InsecureSkipVerify: false,
					}
					// c.ch.tlsConfigDoubleWay.BuildNameToCertificate()

					// tls config single way
					c.ch.tlsConfigSingleWay = &tls.Config{
						Certificates:       []tls.Certificate{cert},
						InsecureSkipVerify: false,
					}
					c.socket.Close()
					return nil
				}

			}
		}
	}
	return nil
}

// RunOutConn : handler for the socket, for Link()
func (c *ShosetConn) runLinkConn() {
	linkConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "link")
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfigDoubleWay)
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		c.socket = conn
		c.rb.UpdateReader(c.socket)
		c.wb.UpdateWriter(c.socket)
		// c.rb = msg.NewReader(c.socket)
		// c.wb = msg.NewWriter(c.socket)
		// defer conn.Close()

		// receive messages
		for {
			if c.GetRemoteLogicalName() == "" {
				err := c.SendMessage(*linkConfig)
				if err != nil {
					c.ch.logger.Error().Msg("couldn't send linkcfg : " + err.Error())
				}
			}

			err := c.receiveMsg()
			time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.SetRemoteLogicalName("") // reinitialize conn
				conn.Close()
				break
			}
		}

	}
}

// RunJoinConn : handler for the socket, for Join()
func (c *ShosetConn) runJoinConn() {
	joinConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "join") //we create a new message config
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfigDoubleWay) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Millisecond * time.Duration(100))
			c.ch.logger.Error().Msg("join err : " + err.Error())
			continue
		}
		// a connection occured
		c.socket = conn
		c.rb.UpdateReader(c.socket)
		c.wb.UpdateWriter(c.socket)
		// c.rb = msg.NewReader(c.socket)
		// c.wb = msg.NewWriter(c.socket)
		// defer conn.Close()

		// receive messages
		for {
			if c.GetRemoteLogicalName() == "" {
				err := c.SendMessage(*joinConfig)
				if err != nil {
					c.ch.logger.Error().Msg("couldn't send joincfg : " + err.Error())
				}
			}

			err := c.receiveMsg()
			time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.ch.logger.Error().Msg("join err in recvmsg : " + err.Error())
				c.SetRemoteLogicalName("") // reinitialize conn
				conn.Close()
				break
			}
		}

	}
}

// runByeConn : handler for the socket, for Bye()
func (c *ShosetConn) runByeConn() {
	byeConfig := msg.NewCfg(c.ch.bindAddress, c.ch.lName, c.ch.ShosetType, "bye") //we create a new message config
	for {
		if !c.GetIsValid() { // sockets are not from the same type or don't have the same name / conn ended
			break
		}

		conn, err := tls.Dial("tcp", c.GetRemoteAddress(), c.ch.tlsConfigDoubleWay) // we wait for a socket to connect each loop

		if err != nil { // no connection occured
			time.Sleep(time.Millisecond * time.Duration(100))
			continue
		}
		// a connection occured
		c.socket = conn
		c.rb.UpdateReader(c.socket)
		c.wb.UpdateWriter(c.socket)
		// c.rb = msg.NewReader(c.socket)
		// c.wb = msg.NewWriter(c.socket)
		// defer conn.Close()

		// receive messages
		for {
			if c.GetRemoteLogicalName() == "" {
				err := c.SendMessage(*byeConfig)
				if err != nil {
					c.ch.logger.Error().Msg("couldn't send byecfg : " + err.Error())
				}
			}

			err := c.receiveMsg()
			time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.SetRemoteLogicalName("") // reinitialize conn
				conn.Close()
				break
			}
		}

	}
}

func (c *ShosetConn) runInConnSingle(address_ string) {
	c.rb.UpdateReader(c.socket)
	c.wb.UpdateWriter(c.socket)
	// c.rb = msg.NewReader(c.socket)
	// c.wb = msg.NewWriter(c.socket)
	// defer c.socket.Close()

	// delete(c.ch.ConnsSingle, address_)

	// receive messages
	msgType, err := c.rb.ReadString()
	msgType = strings.Trim(msgType, "\n")
	// time.Sleep(time.Millisecond * time.Duration(10))
	runtime.Gosched()
	if err != nil {
		// fmt.Println(c.ch.GetPkiRequestAddress(), "## error : ", err)
		return
	}
	fGet, ok := c.ch.Get[msgType]
	if !ok {
		c.ch.logger.Error().Msg("not ok")
		return
	}
	msgVal, err := fGet(c)
	if err != nil {
		c.ch.logger.Error().Msg("didn't find function to handle event : " + err.Error())
		return
	}
	if cmd := msgVal.GetMsgType(); cmd != "pkievt" {
		// linkProtocol := msgVal.(msg.ConfigProtocol)
		c.ch.logger.Error().Msg("not the right cmd : " + cmd)
		// c.socket.Close() ///////////////////////////////////////////
		// fmt.Println("-------")
		// fmt.Println(linkProtocol.GetCommandName())
		// fmt.Println(linkProtocol.GetAddress())
		// fmt.Println("-------")
		// descr := fmt.Sprintf("ConnsByName : ")
		// for _, lName := range c.ch.ConnsByName.Keys() {
		// 	c.ch.ConnsByName.Iterate(lName,
		// 		func(key string, val *ShosetConn) {
		// 			descr = fmt.Sprintf("%s %s\n\t\t\t     ", descr, val)
		// 		})
		// }
		// fmt.Println(descr)
		// fmt.Println("-------")
		// fmt.Println(c.ch.ConnsSingle)
		// fmt.Println("-------")
		// fmt.Println(c.ch.ConnsSingleAddress)
		return
	}
	evt := msgVal.(msg.PkiEvent)
	dirname, err := os.UserHomeDir()
	if err != nil {
		c.ch.logger.Error().Msg("couldn't get dirname : " + err.Error())
		return
	}
	if !c.ch.GetIsPki() {
		// 2. un nouveau se connecte à moi et je suis passe plat
		c.ch.ConnsSingleAddress.Set(evt.GetRequestAddress(), c)
		SendPkiEvent(c.ch, msgVal)
		// c.socket.Close()
		c.ch.ConnsSingle.Delete(address_)
		return
	}
	// 1. un nouveau se connecte directement à moi et je suis PKI
	if evt.GetCertReq() != nil {
		CAcert, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/CAcert.crt")
		if err != nil {
			c.ch.logger.Error().Msg("couldn't get CAcert : " + err.Error())
			return
		}
		signedCert := c.ch.SignCertificate(evt.GetCertReq(), evt.GetHostPublicKey())
		if signedCert != nil {
			var returnPkiEvent *msg.PkiEvent
			var CAprivateKey *rsa.PrivateKey

			if c.ch.GetLogicalName() == evt.GetLogicalName() { // les clusters deviennent à leur tour pki
				CAprivateKeyBytes, err := ioutil.ReadFile(dirname + "/.shoset/" + c.ch.GetFileName() + "/cert/privateCAKey.key")
				if err != nil {
					c.ch.logger.Error().Msg("couldn't get CAprivateKey : " + err.Error())
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
				CAprivateKey, err = x509.ParsePKCS1PrivateKey(b)
				if err != nil {
					fmt.Println(err)
				}
			}
			returnPkiEvent = msg.NewPkiEventReturn("return_pkievt", evt.GetRequestAddress(), signedCert, CAcert, CAprivateKey)
			returnPkiEvent.SetUUID(evt.GetUUID() + "*") // return event has the same uuid so that network isn't flooded with same events
			err := c.SendMessage(*returnPkiEvent)
			if err != nil {
				c.ch.logger.Error().Msg("couldn't send singleConn returnpkievt : " + err.Error())
			}
			c.socket.Close()
			c.ch.ConnsSingle.Delete(address_)
			return
		}
	}
	// 3. j'ai reçu un message autre que pkievt, donc j'ignore
}

// runInConnDouble : handler for the connection, for handleBind()
func (c *ShosetConn) runInConnDouble() {
	c.rb.UpdateReader(c.socket)
	c.wb.UpdateWriter(c.socket)
	// c.rb = msg.NewReader(c.socket)
	// c.wb = msg.NewWriter(c.socket)

	defer c.socket.Close()

	// receive messages
	for {
		err := c.receiveMsg()
		// time.Sleep(time.Millisecond * time.Duration(10))
		runtime.Gosched()
		if err != nil {
			c.ch.logger.Error().Msg("err in receivmsg runInConnDouble : " + err.Error())
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
			return nil
		} else if errors.Is(err, syscall.ECONNRESET) {
			return nil
		}
		c.ch.logger.Warn().Msg("couldn't write string msg : " + err.Error())
		return err
	}
	err = c.WriteMessage(msg)
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			// https://gosamples.dev/broken-pipe/
			return nil
		} else if errors.Is(err, syscall.ECONNRESET) {
			// https://gosamples.dev/connection-reset-by-peer/
			return nil
		}
		c.ch.logger.Warn().Msg("couldn't write message msg : " + err.Error())
		return err
	}
	return nil
}

func (c *ShosetConn) receiveMsg() error {
	if !c.GetIsValid() {
		c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		return errors.New("error : Invalid connection for join - not the same type/name or shosetConn ended")
	}

	// read message type
	msgType, err := c.rb.ReadString()
	switch {
	case err == io.EOF:
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New(err.Error())
	case err != nil:
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New(err.Error())
	}
	msgType = strings.Trim(msgType, "\n")

	if msgType == "hello double" {
		return nil
	}

	// read Message Value
	fGet, ok := c.ch.Get[msgType]
	if ok {
		msgVal, err := fGet(c)
		if err == nil {
			// read message data and handle it with the proper function
			fHandle, ok := c.ch.Handle[msgType]
			if ok {
				go fHandle(c, msgVal)
			}
		} else {
			if c.GetDir() == "in" {
				c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
			}
			return errors.New("receiveMsg : can not read value of " + msgType)
		}
	}
	if !ok {
		if c.GetDir() == "in" {
			c.ch.deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("receiveMsg : non implemented type of message " + msgType)
	}
	time.Sleep(time.Millisecond * time.Duration(100)) // maybe we can remove this sleep time
	return nil
}
