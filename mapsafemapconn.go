package shoset

import (
	"sync"
	"github.com/spf13/viper"
	// "fmt"
)

// MapSafeMapConn : simple key map safe for goroutines...
type MapSafeMapConn struct {
	m map[string]*MapSafeConn
	sync.Mutex
	ConfigName string
	viperConfig *viper.Viper
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

func (m *MapSafeMapConn) GetConfig() []string {
	m.Lock()
	defer m.Unlock()
	return m.viperConfig.GetStringSlice(m.ConfigName)
}

// Set : assign a value to a MapSafeMapConn
func (m *MapSafeMapConn) Set(lname, key string, value *ShosetConn) *MapSafeMapConn {
	m.Lock()
	defer m.Unlock()
	if lname != "" && key != "" {
		if m.m[lname] == nil {
			m.m[lname] = NewMapSafeConn()
		}
		m.m[lname].Set(key, value)
	}

	keys := m.m[lname].Keys()
	if m.ConfigName != "" {
		m.viperConfig.Set(m.ConfigName, keys)
		m.viperConfig.WriteConfigAs("./"+m.ConfigName+".yaml")
	}
	return m
}

func (m*MapSafeMapConn) SetConfigName(name string) {
	if name != "" {
		m.ConfigName = name
	}
}

// Delete : delete a value in a MapSafeMapConn
func (m *MapSafeMapConn) Delete(lname, key string) {
	m.Lock()
	_, ok := m.m[lname]
	if ok {
		m.m[lname].Delete(key)
	}
	if m.ConfigName != "" {
		m.viperConfig.Set(m.ConfigName, m.m[lname].Keys())
		m.viperConfig.WriteConfigAs("./"+m.ConfigName+".yaml")
	}
	m.Unlock()
}

// Iterate : iterate through MapSafeMapConn Values using a function
func (m *MapSafeMapConn) Iterate(lname string, iter func(string, *ShosetConn)) {
	m.Lock()
	mapConn := m.m[lname]
	if mapConn != nil {
		mapConn.Iterate(iter)
	}
	m.Unlock()
}

// Len : return length of the map
func (m *MapSafeMapConn) Len() int {
	return len(m.m)
}

func (m *MapSafeMapConn) SetViper(viperConfig *viper.Viper) {
	m.viperConfig = viperConfig
}
