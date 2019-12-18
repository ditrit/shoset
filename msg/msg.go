package msg

import (
	"time"

	uuid "github.com/kjk/betterguid"
)

// Command : gandalf commands
type Command struct {
	RoutingStep string

	UUID      string
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
	c.UUID = uuid.New()
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

	UUID      string
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

	r.UUID = c.UUID
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
	UUID      string
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
	e.UUID = uuid.New()
	e.Token = token
	e.Timestamp = time.Now().Unix()
	e.Timeout = 1000

	e.Topic = topic
	e.Event = event
	e.Payload = payload
	return e
}

// Config : Gandalf Socket config
type Config struct {
	Name string
}

// NewConfig : Config constructor
func NewConfig(name string) *Config {
	c := new(Config)
	c.Name = name
	return c
}
