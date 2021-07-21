package shoset

import (
	"sync"

	"github.com/spf13/viper"
	"fmt"
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
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", protocolType)
	// fmt.Println("Address will be set with lname-key : ", lname, key)
	if lname != "" && key != "" {
		// fmt.Println("enter condition")
		if m.m[lname] == nil {
			m.m[lname] = NewMapSafeConn()
		}
		m.m[lname].Set(key, value)

	}
	// fmt.Println(m.m[lname])

	keys := m.m[lname].Keys("out")
	// fmt.Println("keys ok : ", keys)
	if m.ConfigName != "" {
		m.viperConfig.Set(protocolType, keys) //temporary - find a better way when link option
		m.viperConfig.WriteConfigAs("./" + m.ConfigName + ".yaml")
	}

	// fmt.Println("Address will be set")
	fmt.Println("!!!!!!!!!!!!!!!okkkkkkkkkkkkkk!!!!!!!!!!!!!!!!!!!!!!!!!!", protocolType)
	return m
}

// Delete : delete a value in a MapSafeMapConn
func (m *MapSafeMapConn) Delete(lname, key, protocolType string) {
	m.Lock()
	_, ok := m.m[lname]
	if ok {
		m.m[lname].Delete(key)
	}
	keys := m.m[lname].Keys("out") // all before, test with out
	if m.ConfigName != "" {
		m.viperConfig.Set(protocolType, keys) //temporary - find a better way
		m.viperConfig.WriteConfigAs("./" + m.ConfigName + ".yaml")
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
	// fmt.Println("enter keys")
	// fmt.Println(m)
	lName := make([]string, m.Len())
	// fmt.Println("lName talble ok")
	i := 0
	for key := range m.m {
		lName[i] = key
		i++
	}
	return lName[:i]
}
