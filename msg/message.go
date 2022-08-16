package msg

import (
	"time"

	uuid "github.com/kjk/betterguid"
)

// Message interface
type Message interface {
	GetMessageType() string
	GetUUID() string
	GetTenant() string
	GetToken() string
	GetTimestamp() int64
	GetTimeout() int64
	GetPayload() string
	GetMajor() int8
	GetMinor() int8
	
	GetDestinationLname() string
}

// MessageBase base struct for messages
type MessageBase struct {
	UUID             string // automatically generated id for a single message
	Tenant           string // ??
	Token            string // ??
	Timeout          int64 // time set to keep event in memory
	Timestamp        int64 // time when message is created
	Payload          string
	Next             string // ??
	Major            int8 // ??
	Minor            int8 // ??
	DestinationLname string // LogicalName the message is destined to
}

// InitMessageBase constructor
func (m *MessageBase) InitMessageBase() {
	m.UUID = uuid.New()
	m.Timestamp = time.Now().UnixMilli() //.Unix()
	m.Timeout = 10
	m.Major = 1
	m.Minor = 0
}

// GetUUID returns UUID from MessageBase.
func (m MessageBase) GetUUID() string { return m.UUID }

// SetUUID sets UUID for MessageBase.
func (m *MessageBase) SetUUID(newUUID string) { m.UUID = newUUID }

// GetTenant returns Tenant from MessageBase.
func (m MessageBase) GetTenant() string { return m.Tenant }

// GetToken returns Token from MessageBase.
func (m MessageBase) GetToken() string { return m.Token }

// GetTimestamp returns Timestamp from MessageBase.
func (m MessageBase) GetTimestamp() int64 { return m.Timestamp }

// GetTimeout returns Timeout from MessageBase.
func (m MessageBase) GetTimeout() int64 { return m.Timeout }

// GetPayload returns Payload from MessageBase.
func (m MessageBase) GetPayload() string { return m.Payload }

// GetMajor returns Major from MessageBase.
func (m MessageBase) GetMajor() int8 { return m.Major }

// GetMinor returns Minor from MessageBase.
func (m MessageBase) GetMinor() int8 { return m.Minor }



// GetDestinationLname returns DestinationLname from MessageBase.
func (m MessageBase) GetDestinationLname() string { return m.DestinationLname }

// Set DestinationLname sets DestinationLname from MessageBase.
func (m *MessageBase) SetDestinationLname(s string) { m.DestinationLname=s }
