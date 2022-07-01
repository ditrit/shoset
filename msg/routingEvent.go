package msg

// RoutingEvent : to broadcast routes between logical names in the network
type RoutingEvent struct {
	Event
	Origin  string
	NbSteps int
}

// NewRoutingEvent : RoutingEvent constructor
func NewRoutingEvent(origin string) *RoutingEvent {
	r := new(RoutingEvent)
	r.Event = *NewEvent(map[string]string{"event": "routing"})

	r.Origin = origin
	r.NbSteps = 1
	return r
}

// GetMsgType accessor
func (r RoutingEvent) GetMsgType() string { return "routingEvent" }

// GetOrigin accessor
func (r RoutingEvent) GetOrigin() string { return r.Origin }

// GetNbSteps accessor
func (r RoutingEvent) GetNbSteps() int { return r.NbSteps }

// GetNb_steps accessor
func (r *RoutingEvent) SetNbSteps(i int) { r.NbSteps = i }
