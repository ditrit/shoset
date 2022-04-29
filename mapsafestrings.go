package shoset

import (
	"sync"
)

// MapSafeStrings : simple key map safe for goroutines...
type MapSafeStrings struct {
	m map[string]map[string]bool
	sync.Mutex
}

// NewMapSafeStrings : constructor
func NewMapSafeStrings() *MapSafeStrings {
	m := new(MapSafeStrings)
	m.m = make(map[string]map[string]bool)
	return m
}

// Get : Get a value from a MapSafeStrings
func (m *MapSafeStrings) Get(key string) map[string]bool {
	m.Lock()
	defer m.Unlock()
	return m.m[key]
}

func (m *MapSafeStrings) Set(key, value string) {
	m.Lock()
	defer m.Unlock()
	if m.m[key] == nil {
		m.m[key] = make(map[string]bool)
	}
	m.m[key][value] = true
}

// Delete : delete a value in a MapSafeStrings
func (m *MapSafeStrings) Delete(key string) {
	m.Lock()
	_, ok := m.m[key]
	if ok {
		delete(m.m, key)
	}
	m.Unlock()
}

// Iterate : iterate through MapSafeStrings Values using a function
func (m *MapSafeStrings) Iterate(iter func(string, map[string]bool)) {
	m.Lock()
	for key, val := range m.m {
		iter(key, val)
	}
	m.Unlock()
}

// Len : return length of the map
func (m *MapSafeStrings) Len() int {
	return len(m.m)
}

func (m *MapSafeStrings) Keys(key string) []string {
	m.Lock()
	defer m.Unlock()
	return m._keys(key)
}

func (m *MapSafeStrings) _keys(key string) []string {
	lNamesByType := m.m[key]
	lNames := make([]string, m.Len())
	i := 0
	for lName := range lNamesByType {
		lNames[i] = lName
		i++
	}
	return lNames[:i]
}
