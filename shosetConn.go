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

// status of the conn
type ProtectedStatus struct {
	value bool
	m     sync.RWMutex
}

// ShosetConn : secured connection based on tls.Conn but with upgraded features
type ShosetConn struct {
	Logger zerolog.Logger // pretty logger

	conn *tls.Conn // secured connection between client and server

	shoset *Shoset // network socket but with upgraded features

	rb *msg.Reader // reader safe for goroutines
	wb *msg.Writer // writer safe for goroutines

	remoteLname      string // logical name of the socket in front of this one
	remoteShosetType string // shosetType of the socket in front of this one
	direction        string // direction of the connection (in or out)
	remoteAddress    string // address of the socket in front of this one

	protocol string // protocol type used by the ShosetConn (join, link, ...) (Usualy is not known ("") at the time of creation of the ShosetConn.)

	isValid ProtectedStatus // status of the ShosetConn
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

// GetDirection returns direction from ShosetConn.
func (c *ShosetConn) GetDirection() string { return c.direction }

// GetRemoteAddress returns remoteAddress from ShosetConn.
func (c *ShosetConn) GetRemoteAddress() string { return c.remoteAddress }

// GetProtocol returns protocol from ShosetConn.
func (c *ShosetConn) GetProtocol() string { return c.protocol }

// GetLocalAddress returns shoset.GetBindAddress() from ShosetConn.
func (c *ShosetConn) GetLocalAddress() string { return c.GetShoset().GetBindAddress() }

// GetIsValid returns isValid from ShosetConn.
func (c *ShosetConn) GetIsValid() bool {
	c.isValid.m.RLock()
	defer c.isValid.m.RUnlock()
	return c.isValid.value
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

// SetProtocol sets the protocol for a ShosetConn.
func (c *ShosetConn) SetProtocol(protocol string) { c.protocol = protocol }

// SetIsValid sets the state for a ShosetConn.
func (c *ShosetConn) SetIsValid(state bool) {
	c.isValid.m.Lock()
	defer c.isValid.m.Unlock()
	c.isValid.value = state
}

// SetRemoteAddress sets the address for a ShosetConn.
func (c *ShosetConn) SetRemoteAddress(address string) { c.remoteAddress = address }

// Stores stores info about ShosetConn and Shoset for protocols
func (c *ShosetConn) Store(protocol, lName, address, shosetType string) {
	c.SetProtocol(protocol)

	c.SetRemoteLogicalName(lName)
	c.SetRemoteShosetType(shosetType)

	c.GetShoset().LnamesByProtocol.AppendToKeys(protocol, lName, true)
	c.GetShoset().LnamesByType.AppendToKeys(shosetType, lName, true)
	c.GetShoset().ConnsByLname.StoreConfig(lName, address, protocol, c)

	// Reroute the network
	routing := msg.NewRoutingEvent(c.GetLocalLogicalName(), true, 0, "")
	c.GetShoset().Send(routing)

	c.SetIsValid(true)
}

// NewShosetConn creates a new ShosetConn object for a specific address.
// Initializes each fields.
func NewShosetConn(s *Shoset, address, direction string) (*ShosetConn, error) {
	ipAddress, err := GetIP(address)
	if err != nil {
		return nil, err
	}

	logger := s.Logger.With().Str("scid", uuid.New()).Logger()
	logger.Debug().Strs("address-direction", []string{address, direction}).Msg("shosetConn created")

	return &ShosetConn{
		Logger:        logger,
		shoset:        s,
		direction:     direction,
		conn:          new(tls.Conn),
		rb:            new(msg.Reader),
		wb:            new(msg.Writer),
		remoteAddress: ipAddress,
		isValid:       ProtectedStatus{value: false},
	}, nil
}

// String returns the formatted string of ShosetConn object in a pretty way.
func (c *ShosetConn) String() string {
	return fmt.Sprintf("ShosetConn{RemoteLogicalName : %s, remoteAddress : %s, type : %s, protocol : %s, way : %s, isValid : %v}", c.GetRemoteLogicalName(), c.GetRemoteAddress(), c.GetRemoteShosetType(), c.GetProtocol(), c.GetDirection(), c.GetIsValid())
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
			time.Sleep(time.Millisecond * time.Duration(2000))
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
	}
}

// ReceiveMessage read incoming message type and runs handleMessageType to handle it.
func (c *ShosetConn) ReceiveMessage() error {
	messageType, err := c.GetReader().ReadString()
	switch {
	case err == io.EOF:
		c.GetShoset().DeleteConn(c.GetRemoteAddress(), c.GetRemoteLogicalName())
		return err
	case errors.Is(err, syscall.ECONNRESET):
		return nil
	case errors.Is(err, syscall.EPIPE):
		return nil
	case err != nil:
		if c.GetDirection() == IN {
			c.GetShoset().DeleteConn(c.GetRemoteLogicalName(), c.GetRemoteAddress())
		}
		return err
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
		if c.GetDirection() == IN {
			c.GetShoset().DeleteConn(c.GetRemoteLogicalName(), c.GetRemoteAddress())
		}
		return errors.New("ReceiveMessage : non implemented type of message " + messageType)
	}

	messageValue, err := handler.Get(c)
	if err != nil {
		if c.GetDirection() == IN {
			c.GetShoset().DeleteConn(c.GetRemoteLogicalName(), c.GetRemoteAddress())
		}
		return errors.New("ReceiveMessage : can not read value of " + messageType + " : " + err.Error())
	}

	// If the message is of a forwardable type, an Acknowledge is expected by the sender
	if contains(FORWARDABLE_TYPES, messageType) {
		// Send back FowarkAck
		forwardAck := msg.NewForwardAck(messageValue.GetUUID(), messageValue.GetTimestamp())
		err := c.GetWriter().SendMessage(forwardAck)

		if err != nil {
			c.Logger.Error().Msg("Couldn't send FowarkAck message : " + err.Error())
		}
	}

	// Check if the destinationLname is the current Lname
	if (messageValue.GetDestinationLname() != c.GetLocalLogicalName()) && messageValue.GetDestinationLname() != "" {
		c.GetShoset().forwardMessage(messageValue)
		return nil
	}

	switch {
	case messageType == TLS_SINGLE_WAY_PKI_EVT:
		err := c.HandleSingleWay(messageValue)
		if err != nil {
			return err
		}
	case contains(MESSAGE_TYPES, messageType):
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
