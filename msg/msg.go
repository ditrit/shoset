package msg

import (
	"time"

	uuid "github.com/kjk/betterguid"
)

// Command : gandalf commands
type Command struct {
	RoutingStep string

	Uuid      string
	Tenant    string
	Token     string
	Timestamp int64
	Timeout   int

	SrcAggregator string
	SrcRouter     string
	SrcWorker     string
	DstAggregator string
	DstRouter     string
	DstWorker     string

	ConnectorType string
	Command       string
	MajorVersion  int8
	MinorVersion  int8
	Payload       string
}

// NewCommand : Command constructor
// todo : passer une map pour gerer les valeurs optionnelles ?
func NewCommand(token string, connectorType string, command string, payload string) *Command {
	c := new(Command)
	c.Uuid = uuid.New()
	c.Token = token
	c.Timestamp = time.Now().Unix()
	c.Timeout = 1000

	c.ConnectorType = connectorType
	c.Command = command
	c.MajorVersion = 1
	c.MinorVersion = 0
	c.Payload = payload
	return c
}

// Reply : gandalf replies to Commands
type Reply struct {
	routingStep string

	Uuid      string
	Tenant    string
	Token     string
	Timestamp int64
	Timeout   int

	SrcAggregator string
	SrcRouter     string
	SrcWorker     string
	DstAggregator string
	DstRouter     string
	DstWorker     string

	State   string
	Payload string
}

// NewReply : Reply constructor
func (c Command) NewReply(state string, payload string) *Reply {
	r := new(Reply)

	r.State = state
	r.Payload = payload

	r.Uuid = c.Uuid
	r.Tenant = c.Tenant
	r.Token = c.Token
	r.Timestamp = c.Timestamp
	r.Timeout = c.Timeout

	r.SrcAggregator = c.SrcAggregator
	r.SrcRouter = c.SrcRouter
	r.SrcWorker = c.SrcWorker
	r.DstAggregator = c.DstAggregator
	r.DstRouter = c.DstRouter
	r.DstWorker = c.DstWorker

	return r
}

// Event : Gandalf events
type Event struct {
	Uuid      string
	Tenant    string
	Token     string
	Timestamp int64
	Timeout   int

	Topic   string
	Event   string
	Payload string
}

// NewEvent : Event constructor
func NewEvent(token string, topic string, event string, payload string) *Event {
	e := new(Event)
	e.Uuid = uuid.New()
	e.Token = token
	e.Timestamp = time.Now().Unix()
	e.Timeout = 1000

	e.Topic = topic
	e.Event = event
	e.Payload = payload
	return e
}

// TODO : SendWith to be replaced by a 'Send' method for the types 'Tcp*'
// func (c Command) SendWith() {
//}

// TODO : from to be replaced by a private method ot the types 'Tcp*'
//func (c Command) from() {
//}

// todo : encode / decode
/*
func (c Command) encodeCommand() (bytesContent []byte, commandError error) {
	bytesContent, err := msgpack.Encode(c)
	if err != nil {
		commandError = fmt.Errorf("Command %s", err)
		return
	}
	return
}

func (c Command) decodeCommand(bytesContent []byte) (command Command, commandError error) {
	err := msgpack.Decode(bytesContent, command)
	if err != nil {
		commandError = fmt.Errorf("Command %s", err)
		return
	}
	return
}

func (cr CommandResponse) encodeCommandResponse() (bytesContent []byte, commandError error) {
	bytesContent, err := msgpack.Encode(cr)
	if err != nil {
		commandError = fmt.Errorf("CommandResponse %s", err)
		return
	}
	return
}

func (cr CommandResponse) decodeCommandResponse(bytesContent []byte) (commandResponse CommandResponse, commandError error) {
	err := msgpack.Decode(bytesContent, commandResponse)
	if err != nil {
		commandError = fmt.Errorf("CommandResponse %s", err)
		return
	}
	return
}
*/
