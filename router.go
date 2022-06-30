package shoset

import (
	//uuid "github.com/kjk/betterguid"
)

type Router struct { // used in map[ destination string] struct{neighbour string, nb_steps int}
	neighbour string // Direction to send message
	nb_steps int // Number of steps to destination logical name
	uuid string

}

func NewRouter(neighbour string, nb_steps int, uuid string) Router {
	return Router{
		neighbour : neighbour,
		nb_steps: nb_steps,
		uuid: uuid,
	}
}

// GetNeighbour returns neighbour from Router.
func (r *Router) GetNeighbour() string { return r.neighbour }

// GetNbSteps returns neighbour from Router.
func (r *Router) GetNbSteps() int { return r.nb_steps }

// GetUUID returns UUID from Router.
func (r *Router) GetUUID() string { return r.uuid }

// SetNeighbour sets neighbour from Router.
func (r *Router) SetNeighbour(s string) { r.neighbour=s }

// SetNbSteps sets nb_steps from Router.
func (r *Router) SetNbSteps(i int) { r.nb_steps=i }

// SetUUID sets UUID from Router.
func (r *Router) SetUUID(s string) { r.uuid=s }

