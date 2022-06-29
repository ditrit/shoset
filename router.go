package shoset

import (
)

type Router struct { // used in map[ destination string] struct{neighbour string, nb_steps int}
	neighbour string // Direction to send message
	nb_steps int // Number of steps to destination logical name
}

func NewRouter(neighbour string, nb_steps int) Router {
	return Router{neighbour : neighbour,nb_steps: nb_steps}
}

// GetNeighbour returns neighbour from Router.
func (r *Router) GetNeighbour() string { return r.neighbour }

// GetNbSteps returns neighbour from Router.
func (r *Router) GetNbSteps() int { return r.nb_steps }

// SetNeighbour sets neighbour from Router.
func (r *Router) SetNeighbour(s string) { r.neighbour=s }

// SetNbSteps sets nb_steps from Router.
func (r *Router) SetNbSteps(i int) { r.nb_steps=i }