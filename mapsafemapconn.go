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

func (m *MapSafeMapConn) GetConfig() []string {
	m.Lock()
	defer m.Unlock()
	return m.viperConfig.GetStringSlice("link") //temporary - find a better way when link option
}

// Set : assign a value to a MapSafeMapConn
func (m *MapSafeMapConn) Set(lname, key string, value *ShosetConn) *MapSafeMapConn {
	m.Lock()
	defer m.Unlock()
	// fmt.Println("Address will be set with lname-key : ", lname, key)
	if lname != "" && key != "" {
		// fmt.Println("enter condition")
		if m.m[lname] == nil {
			m.m[lname] = NewMapSafeConn()
		}
		m.m[lname].Set(key, value)

	}
	// fmt.Println(m.m[lname])

	keys := m.m[lname].Keys()
	// fmt.Println("keys ok : ", keys)
	if m.ConfigName != "" {
		m.viperConfig.Set("bind", keys) //temporary - find a better way when link option
		m.viperConfig.WriteConfigAs("./" + m.ConfigName + ".yaml")
	}

	// fmt.Println("Address will be set")
	return m
}

func (m *MapSafeMapConn) SetConfigName(name string) {
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
	keys := m.m[lname].Keys()
	if m.ConfigName != "" {
		m.viperConfig.Set("bind", keys) //temporary - find a better way when link option
		m.viperConfig.WriteConfigAs("./" + m.ConfigName + ".yaml")
	}
	m.Unlock()
}

// Iterate : iterate through MapSafeMapConn Values using a function
func (m *MapSafeMapConn) Iterate(lnames []string, iter func(string, *ShosetConn)) {
	m.Lock()
	for _, lname := range lnames {
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

func (m *MapSafeMapConn) Keys() []string {
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
