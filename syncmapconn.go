package shoset

import (
	"sync"
)

// SyncMapConn : simple key map safe for goroutines...
type SyncMapConn struct {
	m sync.Map //[string]*ShosetConn
}

// NewSyncMapConn : constructor
func NewSyncMapConn() *SyncMapConn {
	m := new(SyncMapConn)
	// m.m = make(map[string]*ShosetConn)
	return m
}

// Get : Get Map
// func (m *SyncMapConn) GetM() map[string]*ShosetConn {
// 	m.Lock()
// 	defer m.Unlock()
// 	return m.m
// }

// Get : Get a value from a SyncMapConn
// func (m *SyncMapConn) GetByType(shosetType string) []*ShosetConn {
// 	var result []*ShosetConn
// 	m.Lock()
// 	for _, val := range m.m {
// 		if val.GetRemoteShosetType() == shosetType {
// 			result = append(result, val)
// 		}
// 	}
// 	m.Unlock()
// 	return result
// }

// Get : Get a value from a SyncMapConn
func (m *SyncMapConn) Get(key string) *ShosetConn {
	value, err := m.m.Load(key)
	if !err {
		return nil
	}
	return value.(*ShosetConn)
}

// Set : assign a value to a SyncMapConn
func (m *SyncMapConn) Set(key string, value *ShosetConn) *SyncMapConn {
	m.m.Store(key, value)
	return m
}

func (m *SyncMapConn) Keys(dir string) []string { // list of addresses depending on the direction from the conenction wanted
	var addresses []string
	if dir == "all" {
		m.m.Range(func(key, value interface{}) bool {
			addresses = append(addresses, key.(string))
			return true
		})
	} else {
		m.m.Range(func(key, value interface{}) bool {
			if m.Get(key.(string)).GetDir() == dir {
				addresses = append(addresses, key.(string))
			}
			return true
		})
	}
	return addresses
}

// func (m *SyncMapConn) Keys(dir string) []string { // list of addresses
// 	m.Lock()
// 	defer m.Unlock()
// 	return m._keys(dir)
// }

// Delete : delete a value in a SyncMapConn
func (m *SyncMapConn) Delete(key string) {
	m.m.Delete(key)
}

// Iterate : iterate through SyncMapConn Values using a function
func (m *SyncMapConn) Iterate(iter func(string, *ShosetConn)) {
	m.m.Range(func(key, value interface{}) bool {
		iter(key.(string), value.(*ShosetConn))
		return true
	})
}

// Len : return length of the map
// func (m *SyncMapConn) Len() int {
// 	return len(m.m)
// }
