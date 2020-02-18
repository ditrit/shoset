package net

import "sync"

// MapSafe : simple key map safe for goroutines...
type MapSafe struct {
	m map[string]interface{}
	sync.Mutex
}

// NewMapSafe : constructor
func NewMapSafe() *MapSafe {
	m := new(MapSafe)
	m.m = make(map[string]interface{})
	return m
}

// Get : Get a value from a MapSafe
func (m *MapSafe) Get(key string) interface{} {
	m.Lock()
	defer m.Unlock()
	return m.m[key]
}

// Set : assign a value to a MapSafe
func (m *MapSafe) Set(key string, value interface{}) *MapSafe {
	m.Lock()
	m.m[key] = value
	m.Unlock()
	return m
}

// Delete : delete a value in a MapSafe
func (m *MapSafe) Delete(key string) {
	m.Lock()
	_, ok := m.m[key]
	if ok {
		delete(m.m, key)
	}
	m.Unlock()
}

// Iterate : iterate through MapSafe Values using a function
func (m *MapSafe) Iterate(iter func(string, interface{})) {
	m.Lock()
	for key, val := range m.m {
		iter(key, val)
	}
	m.Unlock()
}

// Len : return length of the map
func (m *MapSafe) Len() int {
	return len(m.m)
}
