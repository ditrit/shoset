package shoset

import (
	"fmt"
	"sync"
)

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

// GetM : Get Map
func (m *MapSafeConn) GetM() map[string]*ShosetConn {
	m.Lock()
	defer m.Unlock()
	return m.m
}

// GetByType : Get a value from a MapSafeConn
func (m *MapSafeConn) GetByType(shosetType string) []*ShosetConn {

	var result []*ShosetConn
	m.Lock()
	for _, val := range m.m {
		if val.ShosetType == shosetType {
			result = append(result, val)
		}
	}
	m.Unlock()
	return result
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
func (m *MapSafeConn) Iterate(iter func(string, *ShosetConn) error) error {
	m.Lock()
	defer m.Unlock()
	var e error
	for key, val := range m.m {
		e = iter(key, val)
		if e != nil {
			break
		}
	}
	return e
}

// Len : return length of the map
func (m *MapSafeConn) Len() int {
	return len(m.m)
}

func (m *MapSafeConn) String() string {
	deb := true
	result := "{"
	m.Iterate(func(key string, val *ShosetConn) error {
		if !deb {
			result += ","
		}
		result += fmt.Sprintf(" %s:%s", key, val.String())
		return nil
	})
	result += "} \n"
	return result
}
