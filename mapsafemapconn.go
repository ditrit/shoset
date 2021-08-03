package shoset

import (
	"fmt"
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
func (m *MapSafeMapConn) Set(lname, key, protocolType, shosetType string, value *ShosetConn) *MapSafeMapConn {
	m.Lock()
	defer m.Unlock()
	value.ch.LnamesByProtocol.Set(protocolType, lname)
	value.ch.LnamesByType.Set(shosetType, lname)

	if lname != "" && key != "" {
		if m.m[lname] == nil {
			m.m[lname] = NewMapSafeConn()
		}
		m.m[lname].Set(key, value)
	}
	m.updateFile(lname, protocolType)
	return m
}

// Delete : delete a value in a MapSafeMapConn
func (m *MapSafeMapConn) Delete(lname, key string) {
	m.Lock()
	_, ok := m.m[lname]
	var lNamesByProtocol *MapSafeStrings
	if ok {
		shosetConn := m.m[lname].Get(key)
		if shosetConn != nil {
			lNamesByProtocol = shosetConn.ch.LnamesByProtocol
		}
		m.m[lname].Delete(key)
	}

	if lNamesByProtocol != nil {
		lNamesByProtocol.Iterate(
			func(protocol string, lNames map[string]bool) {
				// if lname in lNames
				// lNamesArray := make([]string, m.Len())
				// i := 0
				// for key := range lNames {
				// 	lNamesArray[i] = key
				// 	i++
				// }
				if lNames[lname] {
					m.updateFile(lname, protocol)
				}
			},
		)
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

func (m *MapSafeMapConn) IterateAll(iter func(string, *ShosetConn)) {
	m.Lock()
	fmt.Println("enter iterateall")
	for _, lname := range m._keys() {
		fmt.Println("key")
		mapConn := m.m[lname]
		if mapConn != nil {
			mapConn.Iterate(iter)
		}
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

func (m *MapSafeMapConn) _keys() []string { // list of logical names inside ConnsByName
	lName := make([]string, m.Len())
	i := 0
	for key := range m.m {
		lName[i] = key
		i++
	}
	return lName[:i]
}

func (m *MapSafeMapConn) Keys() []string { // list of logical names inside ConnsByName
	m.Lock()
	defer m.Unlock()
	return m._keys()
}

func (m *MapSafeMapConn) updateFile(lname, protocolType string) {
	keys := m.m[lname].Keys("out")
	if m.ConfigName != "" && len(keys) != 0 {
		m.viperConfig.Set(protocolType, keys)
		m.viperConfig.WriteConfigAs("./" + m.ConfigName + ".yaml")
	}
}
