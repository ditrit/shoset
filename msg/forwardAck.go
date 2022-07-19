package msg

// SimpleMessage : gandalf commands
type ForwardAck struct {
	MessageBase
	OGMessageUUID      string // UUID of acknowledged message
	OGMessageTimeStamp int64  // TimeStamp of acknowledged message
}

// NewSimpleMessage : ForwardAck constructor
func NewForwardAck(uuid string, timeStamp int64) *ForwardAck {
	var c ForwardAck
	c.InitMessageBase()
	c.OGMessageUUID = uuid
	c.OGMessageTimeStamp = timeStamp
	return &c
}

// GetMessageType accessor
func (c ForwardAck) GetMessageType() string { return "forwardAck" }

// GetOGMessageUUID accessor
func (c ForwardAck) GetOGMessageUUID() string { return c.OGMessageUUID }

// GetOGMessageTimeStamp accessor
func (c ForwardAck) GetOGMessageTimeStamp() int64 { return c.OGMessageTimeStamp }
