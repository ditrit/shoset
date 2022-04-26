package shoset

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

// MapSafeMapConn : simple key map safe for goroutines...
type MapSafeMapConn struct {
	m           sync.Map
	viperConfig *viper.Viper
	sync.Mutex
}

// NewMapSafeMapConn : constructor
func NewMapSafeMapConn() *MapSafeMapConn {
	m := new(MapSafeMapConn)
	return m
}

// Get : Get a value from a MapSafeMapConn
func (m *MapSafeMapConn) Get(key string) *SyncMapConn {
	value, err := m.m.Load(key)
	if !err {
		return nil
	}
	return value.(*SyncMapConn)
}

// func (m *MapSafeMapConn) GetConfig() ([]string, []string) {
// 	// m.Lock()
// 	// defer m.Unlock()
// 	return m._getConfig()
// }

func (m *MapSafeMapConn) GetConfig() ([]string, []string) {
	return m.viperConfig.GetStringSlice("join"), m.viperConfig.GetStringSlice("link")
}

// Set : assign a value to a MapSafeMapConn
func (m *MapSafeMapConn) Set(lname, key, protocolType, shosetType, fileName string, value *ShosetConn) *MapSafeMapConn {
	value.ch.LnamesByProtocol.Set(protocolType, lname)
	value.ch.LnamesByType.Set(shosetType, lname)

	if lname != "" && key != "" {
		if m.Get(lname) == nil {
			m.m.Store(lname, NewSyncMapConn())
		}
		m.Get(lname).Set(key, value)
	}

	keys := m.Get(lname).Keys("out")
	if len(keys) != 0 {
		m.updateFile(lname, protocolType, fileName, keys)
	}
	return m
}

// Delete : delete a value in a MapSafeMapConn
func (m *MapSafeMapConn) Delete(lname, key, fileName string) {
	var lNamesByProtocol *MapSafeStrings
	smc := m.Get(lname)

	if smc != nil {
		shosetConn := smc.Get(key)
		if shosetConn != nil {
			lNamesByProtocol = shosetConn.ch.LnamesByProtocol
		}
		smc.Delete(key)
	}

	var protocolTypes []string
	if lNamesByProtocol != nil {
		lNamesByProtocol.Iterate(
			func(protocol string, lNames map[string]bool) {
				// if lname in lNames
				if lNames[lname] && protocol == "bye" {
					remotesToJoin, remotesToLink := m.GetConfig()
					if contains(remotesToJoin, key) {
						protocolTypes = append(protocolTypes, "join")
					}
					if contains(remotesToLink, key) {
						protocolTypes = append(protocolTypes, "link")
					}

					keys := smc.Keys("out")
					for _, protocolType := range protocolTypes {
						m.updateFile(lname, protocolType, fileName, keys)
					}

				}
			},
		)
	}
}

func (m *MapSafeMapConn) updateFile(lname, protocolType, fileName string, keys []string) {
	m.Lock()
	defer m.Unlock()

	if fileName != "" {
		m.viperConfig.Set(protocolType, keys)
		dirname, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
		}
		m.viperConfig.WriteConfigAs(dirname + "/.shoset/" + fileName + "/config/config.yaml")
	}
}

// Iterate : iterate through MapSafeMapConn Values using a function
func (m *MapSafeMapConn) Iterate(lname string, iter func(string, *ShosetConn)) {
	mapConn := m.Get(lname)
	if mapConn != nil {
		mapConn.Iterate(iter)
	}
}

func (m *MapSafeMapConn) IterateAll(iter func(string, *ShosetConn)) {
	for _, lname := range m.Keys() {
		mapConn := m.Get(lname)
		if mapConn != nil {
			mapConn.Iterate(iter)
		}
	}
}

// Len : return length of the map
// func (m *MapSafeMapConn) Len() int {
// 	return len(m.m)
// }

func (m *MapSafeMapConn) SetViper(viperConfig *viper.Viper) {
	m.viperConfig = viperConfig
}

func (m *MapSafeMapConn) Keys() []string { // list of logical names inside ConnsByName
	var lNames []string
	m.m.Range(func(key, value interface{}) bool {
		lNames = append(lNames, key.(string))
		return true
	})
	return lNames
}

// func (m *MapSafeMapConn) Keys() []string { // list of logical names inside ConnsByName
// 	m.Lock()
// 	defer m.Unlock()
// 	return m._keys()
// }
