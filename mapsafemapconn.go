package shoset

import (
	"fmt"
	"os"
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
	m.updateFile(lname, protocolType, "")
	return m
}

// Delete : delete a value in a MapSafeMapConn
func (m *MapSafeMapConn) Delete(lname, key string) {
	m.Lock()
	var address string
	_, ok := m.m[lname]
	var lNamesByProtocol *MapSafeStrings
	if ok {
		shosetConn := m.m[lname].Get(key)
		if shosetConn != nil {
			address = shosetConn.ch.GetBindAddress()
			// fmt.Println(address, " enter delete")
			lNamesByProtocol = shosetConn.ch.LnamesByProtocol
		}
		if address == "127.0.0.1:8004" || address == "127.0.0.1:8002" {
			fmt.Println(address, " delete : ", key)
		}
		m.m[lname].Delete(key)
	}

	if lNamesByProtocol != nil {
		lNamesByProtocol.Iterate(
			func(protocol string, lNames map[string]bool) {
				// if lname in lNames
				if lNames[lname] {
					if address == "127.0.0.1:8004" || address == "127.0.0.1:8002" {
						fmt.Println(address, " update file")
					}
					m.updateFile(lname, protocol, address)
				}
			},
		)
	}
	m.Unlock()
}

func (m *MapSafeMapConn) updateFile(lname, protocolType, address string) {
	keys := m.m[lname]._keys("out")
	if address == "127.0.0.1:8004" || address == "127.0.0.1:8002" {
		fmt.Println("keys : ", keys)
	}
	if m.ConfigName != "" && len(keys) != 0{
		m.viperConfig.Set(protocolType, keys)
		dirname, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		m.viperConfig.WriteConfigAs(dirname + "/config/" + m.ConfigName + ".yaml")
	}
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
	for _, lname := range m._keys() {
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


