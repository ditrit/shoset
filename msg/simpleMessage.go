package msg

// SimpleMessage : gandalf commands
type SimpleMessage struct {
	MessageBase
}

// NewSimpleMessage : SimpleMessage constructor
func NewSimpleMessage(target string, payload string) *SimpleMessage {
	var c SimpleMessage
	c.InitMessageBase()

	c.SetDestinationLname(target)
	c.Payload=payload
	return &c
}

// GetMessageType accessor
func (c SimpleMessage) GetMessageType() string { return "simpleMessage" }
