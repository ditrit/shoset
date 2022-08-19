package msg

import "time"

// RoutingEvent : to broadcast routes between logical names in the network
type RoutingEvent struct {
	MessageBase
	Origin           string
	NbSteps          int
	RerouteTimestamp int64
}

// NewRoutingEvent : RoutingEvent constructor
func NewRoutingEvent(origin string, GenerateTimestamp bool, RerouteTimestamp int64, uuid string) *RoutingEvent {
	r := new(RoutingEvent)
	r.InitMessageBase()

	r.Origin = origin
	r.NbSteps = 1
	if uuid != "" {
		r.SetUUID(uuid)
	}

	if GenerateTimestamp {
		r.SetRerouteTimestamp(time.Now().UnixMilli())
	} else {
		r.SetRerouteTimestamp(RerouteTimestamp)
	}
	return r
}

// GetMsgType accessor
func (r RoutingEvent) GetMessageType() string { return "routingEvent" }

// GetOrigin accessor
func (r RoutingEvent) GetOrigin() string { return r.Origin }

// GetNbSteps accessor
func (r RoutingEvent) GetNbSteps() int { return r.NbSteps }

// SetNbSteps accessor
func (r *RoutingEvent) SetNbSteps(i int) { r.NbSteps = i }

// GetRerouteTimestamp accessor
func (r RoutingEvent) GetRerouteTimestamp() int64 { return r.RerouteTimestamp }

// SetRerouteTimestamp accessor
func (r *RoutingEvent) SetRerouteTimestamp(i int64) { r.RerouteTimestamp = i }
