package net

import "sync"

// MapSafeConn : simple key map safe for goroutines...
type MapSafeConn struct {
	m map[string]*ShosetConn
	sync.Mutex
}

// NewMapSafeConn : constructor
func NewMapSafeConn() *MapSafeConn {
	m := new(MapSafeConn)
	m.m = make(map[string]*ShosetConn)
	return m
}

// Get : Get a value from a MapSafeConn
func (m *MapSafeConn) Get(key string) *ShosetConn {
	m.Lock()
	defer m.Unlock()
	return m.m[key]
}

// Set : assign a value to a MapSafeConn
func (m *MapSafeConn) Set(key string, value *ShosetConn) *MapSafeConn {
	m.Lock()
	m.m[key] = value
	m.Unlock()
	return m
}

// Delete : delete a value in a MapSafeConn
func (m *MapSafeConn) Delete(key string) {
	m.Lock()
	_, ok := m.m[key]
	if ok {
		delete(m.m, key)
	}
	m.Unlock()
}

// Iterate : iterate through MapSafeConn Values using a function
func (m *MapSafeConn) Iterate(iter func(string, *ShosetConn)) {
	m.Lock()
	for key, val := range m.m {
		iter(key, val)
	}
	m.Unlock()
}

// Len : return length of the map
func (m *MapSafeConn) Len() int {
	return len(m.m)
}
