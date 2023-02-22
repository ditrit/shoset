package shoset

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// MapSyncMap : safe for concurrent use by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
type MapSyncMap struct {
	sync.Map         // map[string]*sync.Map
	cfg      *Config // config file handler
}

// GetConfig returns the config.
func (m *MapSyncMap) GetConfig() *Config {
	return m.cfg
}

// SetConfig sets the config with new one.
func (m *MapSyncMap) SetConfig(cfg *Config) {
	m.cfg = cfg
}

// updateFile updates config after being set
func (m *MapSyncMap) updateFile(protocol string, addresses []string) error {
	m.cfg.AppendToKey(protocol, addresses)

	if err := m.cfg.WriteConfig(m.cfg.GetFileName()); err != nil {
		log.Error().Msg("error writing config: " + err.Error())
		return err
	}
	return nil
}

// StoreConfig sets the value for a key.
// It also updates the viper config file to keep new changes up to date.
// Overrides Store from sync.Map
func (m *MapSyncMap) StoreConfig(lName, address, protocol string, conn interface{}) error {
	// Create a new if it doesn't exists
	if syncMap, _ := m.Load(lName); syncMap == nil {
		m.Store(lName, new(sync.Map))
	}

	syncMap, _ := m.Load(lName)
	syncMap.(*sync.Map).Store(address, conn)

	// OUT is because we only handle the IPaddresses we had to protocol on at some point.
	// They are the one we need to manually reconnect if a problem happens.
	syncMap, _ = m.Load(lName)
	addresses := Keys(syncMap.(*sync.Map), ALL) //OUT
	log.Trace().Msg("Storing in config : " + protocol + " keys : " + fmt.Sprint(addresses))
	if len(addresses) != 0 {
		err := m.updateFile(protocol, addresses)
		return err
	}
	return nil
}

// Iterate calls *MapSync.Range(iter) sequentially for each key and *MapSync present in the map.
func (m *MapSyncMap) Iterate(iter func(string, string, interface{})) {
	if m != nil {
		m.Range(func(key, syncMap interface{}) bool {
			if syncMap != nil {
				syncMap.(*sync.Map).Range(func(key2, val2 interface{}) bool {
					iter(key.(string), key2.(string), val2)
					return true
				})
			}
			return true
		})
	}
}

// Keys returns a []string corresponding to the keys from the map[string]*sync.Map object.
// sType set the specific keys depending on the desired shoset type.
func (m *MapSyncMap) Keys(sType string) []string {
	var keys []string
	if sType == ALL {
		// all keys whatever the type from the ShosetConn
		m.Range(func(key, _ interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
	} else {
		// keys with specific ShosetConn type
		m.Range(func(key, _ interface{}) bool {
			if key == sType {
				keys = append(keys, key.(string))
			}
			return true
		})
	}
	return removeDuplicateStr(keys)
}

// Apppend a value to a MapSyncMap : m[key1][key2]=value
func (m *MapSyncMap) AppendToKeys(key1 string, key2 string, value interface{}) {
	if mapSync, ok := m.Load(key1); ok {
		mapSync.(*sync.Map).Store(key2, value)
		m.Store(key1, mapSync)
	} else {
		mapSync := new(sync.Map)
		mapSync.Store(key2, value)
		m.Store(key1, mapSync)
	}
}

func (m *MapSyncMap) String() string {
	description := ""
	if m != nil {
		m.Range(func(key, syncMap interface{}) bool {
			description += fmt.Sprintf("\tkey : %v", key)
			if syncMap != nil {
				syncMap.(*sync.Map).Range(func(key2, val2 interface{}) bool {
					description += fmt.Sprintf("\n\t\t\tkey : %v val : %v", key2, val2)
					return true
				})
			}
			description += "\n"
			return true
		})
	}
	return description
}

// Retrieve a value from a MapSyncMap : m[key1][key2]=value
func (m *MapSyncMap) LoadValueFromKeys(key1 string, key2 string) (interface{}, bool) {
	if syncMap, ok := m.Load(key1); ok {
		return syncMap.(*sync.Map).Load(key2)
	} else {
		return nil, false
	}
}

// Retrieve a value from a MapSyncMap : m[key1][key2]=value
func (m *MapSyncMap) DeleteValueFromKeys(key1 string, key2 string) {
	if syncMap, ok := m.Load(key1); ok {
		syncMap.(*sync.Map).Delete(key2)

		// Delete empty key1
		var i int
		syncMap2, ok := m.Load(key1)
		if ok {
			syncMap2.(*sync.Map).Range(func(k, v interface{}) bool {
				i++
				return i <= 0 // If i is no longer 0 there is no need to continue
			})
			if i == 0 {
				m.Delete(key1)
			}
		}
	}
}
