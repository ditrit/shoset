package msg

// RoutingEvent : to broadcast routes between logical names in the network
type RoutingEvent struct {
	MessageBase
	Origin  string
	NbSteps int
	Dir     string
}

// NewRoutingEvent : RoutingEvent constructor
func NewRoutingEvent(origin, dir string) *RoutingEvent {
	r := new(RoutingEvent)
	r.InitMessageBase()

	r.Origin = origin
	r.NbSteps = 1
	r.Dir= dir
	return r
}

// GetMsgType accessor
func (r RoutingEvent) GetMessageType() string { return "routingEvent" }

// GetOrigin accessor
func (r RoutingEvent) GetOrigin() string { return r.Origin }

// GetNbSteps accessor
func (r RoutingEvent) GetNbSteps() int { return r.NbSteps }

// GetNbSteps accessor
func (r RoutingEvent) GetDir() string { return r.Dir }

// GetNb_steps accessor
func (r *RoutingEvent) SetNbSteps(i int) { r.NbSteps = i }
