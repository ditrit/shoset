package shoset

import (
	"sync"

	"github.com/spf13/viper"
)

// MapSafeMapConn : simple key map safe for goroutines...
type MapSafeMapConn struct {
	m map[string]*MapSafeConn
	sync.Mutex
	ConfigName  string
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

func (m *MapSafeMapConn) GetConfig() ([]string, []string) {
	m.Lock()
	defer m.Unlock()
	return m.viperConfig.GetStringSlice("join"), m.viperConfig.GetStringSlice("link")
}

// Set : assign a value to a MapSafeMapConn
func (m *MapSafeMapConn) Set(lname, key, protocolType string, value *ShosetConn) *MapSafeMapConn {
	m.Lock()
	defer m.Unlock()

	if lname != "" && key != "" {
		if m.m[lname] == nil {
			m.m[lname] = NewMapSafeConn()
		}
		m.m[lname].Set(key, value)
	}

	keys := m.m[lname].Keys("out")
	if m.ConfigName != "" && len(keys) != 0 {
		m.viperConfig.Set(protocolType, keys)
		m.viperConfig.WriteConfigAs("./" + m.ConfigName + ".yaml")
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

func (m *MapSafeMapConn) SetConfigName(name string) {
	if name != "" {
		m.ConfigName = name
	}
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

func (m *MapSafeMapConn) Keys() []string { // list of logical names inside ConnsByName
	lName := make([]string, m.Len())
	i := 0
	for key := range m.m {
		lName[i] = key
		i++
	}
	return lName[:i]
}
