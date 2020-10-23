package msg

import (
	"time"

	uuid "github.com/kjk/betterguid"
)

// Message interface
type Message interface {
	GetMsgType() string
	GetUUID() string
	GetTenant() string
	GetToken() string
	GetTimestamp() int64
	GetTimeout() int64
	GetPayload() string
	GetVersion() float32
	GetMajor() int8
	GetMinor() int8
}

// MessageBase base struct for messages
type MessageBase struct {
	UUID      string
	Tenant    string
	Token     string
	Timeout   int64
	Timestamp int64
	Payload   string
	Next      string
	Version   float32
	Major     int8
	Minor     int8
}

// InitMessageBase constructor
func (m *MessageBase) InitMessageBase() {
	m.UUID = uuid.New()
	m.Timestamp = time.Now().Unix()
	m.Timeout = 1000
	m.Version = 1.0
	m.Major = 1
	m.Minor = 0
}

// GetUUID accessor
func (m MessageBase) GetUUID() string {
	return m.UUID
}

// GetTenant accessor
func (m MessageBase) GetTenant() string {
	return m.Tenant
}

// GetToken accessor
func (m MessageBase) GetToken() string {
	return m.Token
}

// GetTimestamp accessor
func (m MessageBase) GetTimestamp() int64 {
	return m.Timestamp
}

// GetTimeout accessor
func (m MessageBase) GetTimeout() int64 {
	return m.Timeout
}

// GetPayload accessor
func (m MessageBase) GetPayload() string {
	return m.Payload
}

// GetVersion accessor
func (m MessageBase) GetVersion() float32 {
	return m.Version
}

// GetMajor accessor
func (m MessageBase) GetMajor() int8 {
	return m.Major
}

// GetMinor accessor
func (m MessageBase) GetMinor() int8 {
	return m.Minor
}
