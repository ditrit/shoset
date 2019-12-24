package msg

// Reply : gandalf replies to Commands
type Reply struct {
	MessageBase
	State   string
	CmdUUID string
}

// NewReply : Reply constructor
func NewReply(c *Command, state string, payload string) *Reply {
	r := new(Reply)
	r.InitMessageBase()

	r.CmdUUID = c.UUID
	r.State = state
	r.Payload = payload
	return r
}

// GetMsgType accessor
func (r Reply) GetMsgType() string { return "rep" }

// GetState :
func (r Reply) GetState() string { return r.State }

// GetCmdUUID :
func (r Reply) GetCmdUUID() string { return r.CmdUUID }
