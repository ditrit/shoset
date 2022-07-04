package msg

// RoutingEvent : to broadcast routes between logical names in the network
type RoutingEvent struct {
	MessageBase
	Origin  string
	NbSteps int
}

// NewRoutingEvent : RoutingEvent constructor
func NewRoutingEvent(origin, uuid string) *RoutingEvent {
	r := new(RoutingEvent)
	r.InitMessageBase()

	r.Origin = origin
	r.NbSteps = 1
	if uuid != "" {
		r.SetUUID(uuid)
	}
	return r
}

// GetMsgType accessor
func (r RoutingEvent) GetMessageType() string { return "routingEvent" }

// GetOrigin accessor
func (r RoutingEvent) GetOrigin() string { return r.Origin }

// GetNbSteps accessor
func (r RoutingEvent) GetNbSteps() int { return r.NbSteps }

// GetNb_steps accessor
func (r *RoutingEvent) SetNbSteps(i int) { r.NbSteps = i }
