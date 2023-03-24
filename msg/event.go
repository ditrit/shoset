package msg

// Event : Gandalf event internal
type Event struct {
	MessageBase
	Topic         string
	Event         string
	ReferenceUUID string
}

// NewEvent : Event constructor
func NewEvent(params map[string]string) *Event {
	e := new(Event)
	e.InitMessageBase()

	e.Topic = params["topic"]
	e.Event = params["event"]
	e.Payload = params["payload"]
	if val, ok := params["referenceUUID"]; ok {
		e.ReferenceUUID = val
	}
	return e
}

// NewEventClassic : Event Classic constructor
func NewEventClassic(topic, event, payload string) *Event {
	var tab = map[string]string{
		"topic":   topic,
		"event":   event,
		"payload": payload,
	}

	return NewEvent(tab)
}

// GetMessageType accessor
func (e Event) GetMessageType() string { return "evt" }

// GetTopic :
func (e Event) GetTopic() string { return e.Topic }

// GetEvent :
func (e Event) GetEvent() string { return e.Event }

// GetReferenceUUID :
func (e Event) GetReferenceUUID() string { return e.ReferenceUUID }
