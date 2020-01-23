package msg

// Event : Gandalf event internal
type Event struct {
	MessageBase
	Topic          string
	Event          string
	ReferencesUUID string
}

// NewEvent : Event constructor
func NewEvent(params map[string]string) *Event {
	e := new(Event)
	e.InitMessageBase()

	e.Topic = params["topic"]
	e.Event = params["event"]
	e.Payload = params["payload"]
	if val, ok := params["referencesUUID"]; ok {
		e.ReferencesUUID = val
	}
	return e
}

// GetMsgType accessor
func (e Event) GetMsgType() string { return "evt" }

// GetTopic :
func (e Event) GetTopic() string { return e.Topic }

// GetEvent :
func (e Event) GetEvent() string { return e.Event }

// ReferencesUUID :
func (e Event) GetReferencesUUID() string { return e.ReferencesUUID }

