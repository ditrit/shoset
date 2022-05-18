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
	Major     int8
	Minor     int8
}

// InitMessageBase constructor
func (m *MessageBase) InitMessageBase() {
	m.UUID = uuid.New()
	m.Timestamp = time.Now().Unix()
	m.Timeout = 10 // time set to keep event in memory
	m.Major = 1
	m.Minor = 0
}

// GetUUID accessor
func (m MessageBase) GetUUID() string {
	return m.UUID
}

func (m *MessageBase) SetUUID(newUUID string) {
	m.UUID = newUUID
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

// GetMajor accessor
func (m MessageBase) GetMajor() int8 {
	return m.Major
}

// GetMinor accessor
func (m MessageBase) GetMinor() int8 {
	return m.Minor
}
