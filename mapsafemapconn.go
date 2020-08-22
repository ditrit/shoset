package shoset

import (
	"fmt"
	"sync"
)

// MapSafeMapConn : simple key map safe for goroutines...
type MapSafeMapConn struct {
	m map[string]*MapSafeConn
	sync.Mutex
}

// NewMapSafeMapConn : constructor
func NewMapSafeMapConn() *MapSafeMapConn {
	m := new(MapSafeMapConn)
	m.m = make(map[string]*MapSafeConn)
	return m
}

// Get : Get a value from a MapSafeMapConn
func (m *MapSafeMapConn) Get(key string) *MapSafeConn {
	m.Lock()
	defer m.Unlock()
	return m.m[key]
}

// Set : assign a value to a MapSafeMapConn
func (m *MapSafeMapConn) Set(lname string, key string, value *ShosetConn) *MapSafeMapConn {
	if lname != "" && key != "" {
		m.Lock()
		if m.m[lname] == nil {
			m.m[lname] = NewMapSafeConn()
		}
		m.m[lname].Set(key, value)
		m.Unlock()
	}
	return m
}

// Delete : delete a value in a MapSafeMapConn
func (m *MapSafeMapConn) Delete(lname, key string) {
	m.Lock()
	_, ok := m.m[lname]
	if ok {
		m.m[lname].Delete(key)
	}
	m.Unlock()
}

// Iterate : iterate through MapSafeMapConn Values using a function
func (m *MapSafeMapConn) Iterate(lname string, iter func(string, *ShosetConn) error) error {
	m.Lock()
	defer m.Unlock()
	mapConn := m.m[lname]
	if mapConn != nil {
		return mapConn.Iterate(iter)
	}
	return nil
}

// Len : return length of the map
func (m *MapSafeMapConn) Len() int {
	return len(m.m)
}

func (m *MapSafeMapConn) String() string {
	deb := true
	result := "{"
	m.Lock()
	for key, val := range m.m {
		if !deb {
			result += ","
		}
		result += fmt.Sprintf(" %s:%s", key, val.String())
	}
	m.Unlock()
	result += "} \n"
	return result
}
