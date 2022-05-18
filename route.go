package shoset

type Route struct {
	neighbour    string      // Direction to send message
	neighborConn *ShosetConn // ShosetConn to the neighbour
	nb_steps     int         // Number of steps to destination logical name
	uuid         string      // UUID of the message that broadcasted this route
	timestamp    int64
}

func NewRoute(neighbour string, c *ShosetConn, nb_steps int, uuid string, timestamp int64) Route {
	return Route{
		neighbour:    neighbour,
		neighborConn: c,
		nb_steps:     nb_steps,
		uuid:         uuid,
		timestamp:    timestamp,
	}
}

// GetNeighbour returns neighbour from Route.
func (r Route) GetNeighbour() string { return r.neighbour }

// GetneighborConn returns ShosetConn to the neighbour from Route.
func (r Route) GetNeighborConn() *ShosetConn { return r.neighborConn }

// GetNbSteps returns the number of steps of the Route.
func (r Route) GetNbSteps() int { return r.nb_steps }

// GetUUID returns the UUID from the routing event that created this Route.
func (r Route) GetUUID() string { return r.uuid }

// GetTimestamp returns timestamp of the creation of the routing event that created this Route.
func (r Route) GetTimestamp() int64 { return r.timestamp }

// SetNeighbour sets neighbour from Route.
func (r *Route) SetNeighbour(s string) { r.neighbour = s }

// SetNeighborConn sets NeighborConn from Route.
func (r *Route) SetNeighborConn(c *ShosetConn) { r.neighborConn = c }

// SetNbSteps sets nb_steps from Route.
func (r *Route) SetNbSteps(i int) { r.nb_steps = i }
