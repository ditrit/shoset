package msg

// Event : Gandalf event internal
type Event struct {
	MessageBase
	Topic string
	Event string
}

// NewEvent : Event constructor
func NewEvent(topic string, event string, payload string) *Event {
	e := new(Event)
	e.InitMessageBase()

	e.Topic = topic
	e.Event = event
	e.Payload = payload
	return e
}

// GetMsgType accessor
func (e Event) GetMsgType() string { return "evt" }

// GetTopic :
func (e Event) GetTopic() string { return e.Topic }

// GetEvent :
func (e Event) GetEvent() string { return e.Event }
