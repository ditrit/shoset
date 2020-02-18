package net

import "sync"

// MapSafeBool : simple key map safe for goroutines...
type MapSafeBool struct {
	m map[string]bool
	sync.Mutex
}

// NewMapSafeBool : constructor
func NewMapSafeBool() *MapSafeBool {
	m := new(MapSafeBool)
	m.m = make(map[string]bool)
	return m
}

// Get : Get a value from a MapSafeBool
func (m *MapSafeBool) Get(key string) bool {
	m.Lock()
	defer m.Unlock()
	return m.m[key]
}

// Set : assign a value to a MapSafeBool
func (m *MapSafeBool) Set(key string, value bool) *MapSafeBool {
	m.Lock()
	m.m[key] = value
	m.Unlock()
	return m
}

// Delete : delete a value in a MapSafeBool
func (m *MapSafeBool) Delete(key string) {
	m.Lock()
	_, ok := m.m[key]
	if ok {
		delete(m.m, key)
	}
	m.Unlock()
}

// Iterate : iterate through MapSafeBool Values using a function
func (m *MapSafeBool) Iterate(iter func(string, bool)) {
	m.Lock()
	for key, val := range m.m {
		iter(key, val)
	}
	m.Unlock()
}

// Len : return length of the map
func (m *MapSafeBool) Len() int {
	return len(m.m)
}
