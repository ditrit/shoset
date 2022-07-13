package shoset

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"

	// "runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ditrit/shoset/msg"
	uuid "github.com/kjk/betterguid"
	"github.com/rs/zerolog"
)

type ProtectedStatus struct {
	value bool
	m     sync.Mutex
} // state of the conn

// ShosetConn : secured connection based on tls.Conn but with upgraded features
type ShosetConn struct {
	Logger zerolog.Logger // pretty logger

	conn *tls.Conn // secured connection between client and server

	shoset *Shoset // network socket but with upgraded features

	rb *msg.Reader // reader safe for goroutines
	wb *msg.Writer // writer safe for goroutines

	remoteLname      string // logical name of the socket in front of this one
	remoteShosetType string // shosetType of the socket in front of this one
	dir              string
	remoteAddress    string // address of the socket in front of this one

	isValid ProtectedStatus // state of the conn

	StatusChange chan bool // Send a message on the channel to notifie waiting GoRoutines that the state of the connexion Changed (Actual value sent is not used)
}

// GetConn returns conn from ShosetConn.
func (c *ShosetConn) GetConn() *tls.Conn { return c.conn }

// GetShoset returns shoset from ShosetConn.
func (c *ShosetConn) GetShoset() *Shoset { return c.shoset }

// GetReader returns rb from ShosetConn.
func (c *ShosetConn) GetReader() *msg.Reader { return c.rb }

// GetWriter returns wb from ShosetConn.
func (c *ShosetConn) GetWriter() *msg.Writer { return c.wb }

// GetRemoteLogicalName returns remoteLname from ShosetConn.
func (c *ShosetConn) GetRemoteLogicalName() string { return c.remoteLname }

// GetLocalLogicalName returns shoset.GetLogicalName() from ShosetConn.
func (c *ShosetConn) GetLocalLogicalName() string { return c.GetShoset().GetLogicalName() }

// GetRemoteShosetType returns remoteShosetType from ShosetConn.
func (c *ShosetConn) GetRemoteShosetType() string { return c.remoteShosetType }

// GetLocalShosetType returns shoset.GetShosetType() from ShosetConn.
func (c *ShosetConn) GetLocalShosetType() string { return c.GetShoset().GetShosetType() }

// GetDir returns dir from ShosetConn.
func (c *ShosetConn) GetDir() string { return c.dir }

// GetRemoteAddress returns remoteAddress from ShosetConn.
func (c *ShosetConn) GetRemoteAddress() string { return c.remoteAddress }

// GetLocalAddress returns shoset.GetBindAddress() from ShosetConn.
func (c *ShosetConn) GetLocalAddress() string { return c.GetShoset().GetBindAddress() }

// GetIsValid returns isValid from ShosetConn.
func (c *ShosetConn) GetIsValid() bool {
	//c.isValid.m.Lock()
	//defer c.isValid.m.Unlock() // Create lockup
	return c.isValid.value

}

// Wait for ShosetConn to be ready for use
func (c *ShosetConn) WaitForValid() {
	for {
		c.isValid.m.Lock()
		value := c.isValid.value
		c.isValid.m.Unlock()
		if !value {
			<-c.StatusChange
		} else {
			return
		}
	}
}

// SetConn sets the lName for a ShosetConn.
func (c *ShosetConn) SetConn(conn *tls.Conn) { c.conn = conn }

// UpdateConn updates conn attribute along with its reader and writer.
func (c *ShosetConn) UpdateConn(conn *tls.Conn) {
	c.SetConn(conn)
	c.GetReader().UpdateReader(conn)
	c.GetWriter().UpdateWriter(conn)
}

// SetRemoteLogicalName sets the lName for a ShosetConn.
func (c *ShosetConn) SetRemoteLogicalName(lName string) { c.remoteLname = lName }

// SetLocalAddress sets the bindAddress for a ShosetConn.
func (c *ShosetConn) SetLocalAddress(bindAddress string) { c.GetShoset().SetBindAddress(bindAddress) }

// SetRemoteShosetType sets the ShosetType for a ShosetConn.
func (c *ShosetConn) SetRemoteShosetType(ShosetType string) { c.remoteShosetType = ShosetType }

// SetIsValid sets the state for a ShosetConn.
func (c *ShosetConn) SetIsValid(state bool) {
	//c.isValid.m.Lock()
	//defer c.isValid.m.Unlock()

	c.isValid.m.Lock()
	c.isValid.value = state
	c.isValid.m.Unlock()
	if state {
		fmt.Println(c, "Is ready.")
	}
	select {
	case c.StatusChange <- true:
	default:
		fmt.Println("Nobody is waiting for StateChnage")
	}
}

// SetRemoteAddress sets the address for a ShosetConn.
func (c *ShosetConn) SetRemoteAddress(address string) { c.remoteAddress = address }

// Store stores info concerning ShosetConn and Shoset for protocols
func (c *ShosetConn) Store(protocol, lName, address, shosetType string) {
	c.SetRemoteLogicalName(lName)
	c.SetRemoteShosetType(shosetType)

	mapSync := new(sync.Map)
	mapSync.Store(lName, true)
	c.GetShoset().LnamesByProtocol.Store(protocol, mapSync)
	c.GetShoset().LnamesByType.Store(shosetType, mapSync)
	c.GetShoset().ConnsByLname.StoreConfig(lName, address, protocol, c)
}

// NewShosetConn creates a new ShosetConn object for a specific address.
// Initializes each fields.
func NewShosetConn(s *Shoset, address, dir string) (*ShosetConn, error) {
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}

	logger := s.Logger.With().Str("scid", uuid.New()).Logger()
	logger.Debug().Strs("address-dir", []string{address, dir}).Msg("shosetConn created")

	return &ShosetConn{
		Logger:        logger,
		shoset:        s,
		dir:           dir,
		conn:          new(tls.Conn),
		rb:            new(msg.Reader),
		wb:            new(msg.Writer),
		remoteAddress: ipAddress,
		isValid:       ProtectedStatus{value: false},
	}, nil
}

// String returns the formatted string of ShosetConn object in a pretty way.
func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{name: %s, type: %s, way: %s, remoteAddress: %s, isValid: %v}", c.GetRemoteLogicalName(), c.GetRemoteShosetType(), c.GetDir(), c.GetRemoteAddress(), c.GetIsValid())
}

// HandleConfig handles ConfigProtocol message.
// Connects to the remote address and sends the protocol through this connection.
func (c *ShosetConn) HandleConfig(cfg *msg.ConfigProtocol) {
	defer func() {
		c.Logger.Debug().Msg("HandleConfig: socket closed")
		c.GetConn().Close()
	}()
	for {
		protocolConn, err := tls.Dial(CONNECTION_TYPE, c.GetRemoteAddress(), c.GetShoset().GetTlsConfigDoubleWay())
		if err != nil {
			time.Sleep(time.Millisecond * time.Duration(10))
			c.Logger.Error().Msg("HandleConfig err: " + err.Error())
			continue
		}

		c.UpdateConn(protocolConn)

		err = c.GetWriter().SendMessage(*cfg)
		if err != nil {
			c.Logger.Error().Msg("couldn't send cfg: " + err.Error())
			continue
		}

		for {
			err := c.ReceiveMessage()
			if err != nil {
				c.Logger.Error().Msg("socket closed: err in ReceiveMessage HandleConfig: " + err.Error())
				break
			}
		}

	}
}

// RunInConnSingle runs ReceiveMessage for TLS Single Way connection.
func (c *ShosetConn) RunInConnSingle(address string) {
	c.GetShoset().ConnsSingleBool.Delete(address)

	err := c.ReceiveMessage()
	if err != nil {
		c.Logger.Error().Msg("socket closed: err in ReceiveMessage RunInConnSingle: " + err.Error())
		return
	}
}

// RunInConnDouble runs ReceiveMessage for TLS Double Way connection.
func (c *ShosetConn) RunInConnDouble() {
	defer func() {
		c.Logger.Debug().Msg("double_way: socket closed")
		c.GetConn().Close()
	}()

	for {
		err := c.ReceiveMessage()
		if err != nil {
			c.Logger.Error().Msg("err in ReceiveMessage RunInConnDouble: " + err.Error())
			return
		}

		for {
			err := c.ReceiveMessage()
			// time.Sleep(time.Millisecond * time.Duration(100))
			if err != nil {
				c.Logger.Error().Msg("socket closed: err in ReceiveMessage HandleConfig: " + err.Error())
				break
			}
		}

	}
}

// ReceiveMessage read incoming message type and runs handleMessageType to handle it.
func (c *ShosetConn) ReceiveMessage() error {
	messageType, err := c.GetReader().ReadString()
	switch {
	case err == io.EOF:
		if c.GetDir() == IN {
			c.GetShoset().deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return nil
	case errors.Is(err, syscall.ECONNRESET):
		return nil
	case errors.Is(err, syscall.EPIPE):
		return nil
	case err != nil:
		if c.GetDir() == IN {
			c.GetShoset().deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New(err.Error())
	}
	messageType = strings.Trim(messageType, "\n")

	if messageType == TLS_DOUBLE_WAY_TEST_WRITE { // do not handle this message, test for shoset.handleBind()
		return nil
	}

	err = c.handleMessageType(messageType)
	if err != nil {
		return err
	}
	return nil
}

// handleMessageType deduce handler from messageType and use it adequately.
func (c *ShosetConn) handleMessageType(messageType string) error {
	handler, ok := c.GetShoset().Handlers[messageType]
	if !ok {
		if c.GetDir() == IN {
			c.GetShoset().deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("ReceiveMessage : non implemented type of message " + messageType)
	}

	messageValue, err := handler.Get(c)
	if err != nil {
		if c.GetDir() == IN {
			c.GetShoset().deleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		}
		return errors.New("ReceiveMessage : can not read value of " + messageType + " : " + err.Error())
	}

	// Check if the destinationLname is the current Lname
	if (messageValue.GetDestinationLname() != c.GetLocalLogicalName()) && messageValue.GetDestinationLname() != "" {
		c.GetShoset().forwardMessage(messageValue)
	}

	doubleWayMessageTypes := []string{"cfgjoin", "cfglink", "cfgbye", "pkievt_TLSdoubleWay", "routingEvent", "evt", "cmd", "simpleMessage"} //added "routingEvent", "evt", "cmd", "simpleMessage"
	//fmt.Println("(handleMessageType) messageType : ",messageType)
	switch {
	case messageType == TLS_SINGLE_WAY_PKI_EVT:
		err := c.HandleSingleWay(messageValue)
		if err != nil {
			return err
		}
	case contains(doubleWayMessageTypes, messageType):
		err := handler.HandleDoubleWay(c, messageValue)
		if err != nil {
			return err
		}
	default:
		c.Logger.Error().Msg("wrong messageType : " + messageType)
		return errors.New("wrong messageType : " + messageType)
	}
	return nil
}
