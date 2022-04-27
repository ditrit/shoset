package shoset

import (
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

// Get : Get Map
func (m *MapSafeConn) GetM() map[string]*ShosetConn {
	m.Lock()
	defer m.Unlock()
	return m.m
}

// Get : Get a value from a MapSafeConn
func (m *MapSafeConn) GetByType(shosetType string) []*ShosetConn {
	var result []*ShosetConn
	m.Lock()
	for _, val := range m.m {
		if val.GetRemoteShosetType() == shosetType {
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
	// fmt.Println("Address set")
	m.m[key] = value
	m.Unlock()
	return m
}

func (m *MapSafeConn) _keys(dir string) []string { // list of addresses
	addresses := make([]string, m.Len()+1)
	i := 0
	for key := range m.m {
		if dir != "all" {
			if m.m[key].GetDir() == dir { // on ne veut pas le in du join
				addresses[i] = key
				i++
			}
		} else {
			addresses[i] = key
			i++
		}
	}
	return addresses[:i]
}

func (m *MapSafeConn) Keys(dir string) []string { // list of addresses
	m.Lock()
	defer m.Unlock()
	return m._keys(dir)

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
