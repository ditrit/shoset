package shoset

type Route struct { // used in map[ destination string] struct{neighbour string, nb_steps int}
	neighbour string // Direction to send message
	nb_steps  int    // Number of steps to destination logical name
	uuid      string // UUID of the message that broadcasted this route
	timestamp int64
}

func NewRoute(neighbour string, nb_steps int, uuid string, timestamp int64) Route {
	return Route{
		neighbour: neighbour,
		nb_steps:  nb_steps,
		uuid:      uuid,
		timestamp: timestamp,
	}
}

// GetNeighbour returns neighbour from Route.
func (r Route) GetNeighbour() string { return r.neighbour }

// GetNbSteps returns neighbour from Route.
func (r Route) GetNbSteps() int { return r.nb_steps }

// GetUUID returns UUID from Route.
func (r Route) GetUUID() string { return r.uuid }

// SetNeighbour sets neighbour from Route.
func (r *Route) SetNeighbour(s string) { r.neighbour = s }

// SetNbSteps sets nb_steps from Route.
func (r *Route) SetNbSteps(i int) { r.nb_steps = i }

// SetUUID sets UUID from Route.
func (r *Route) SetUUID(s string) { r.uuid = s }
