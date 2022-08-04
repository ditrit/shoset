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

// SetConfig sets the config with new one.
func (m *MapSyncMap) GetConfig() *Config {
	return m.cfg
}

// SetConfig sets the config with new one.
func (m *MapSyncMap) SetConfig(cfg *Config) {
	m.cfg = cfg
}

func (m *MapSyncMap) prepareUpdateFile(protocol string, ipAddresses []string) {
	var protocols []string

	if contains(m.cfg.GetSlice(PROTOCOL_JOIN), protocol) {
		protocols = append(protocols, PROTOCOL_JOIN)
	}
	if contains(m.cfg.GetSlice(PROTOCOL_LINK), protocol) {
		protocols = append(protocols, PROTOCOL_LINK)
	}

	for _, protocol := range protocols {
		m.updateFile(protocol, ipAddresses)
	}
}

// updateFile updates config after being set
func (m *MapSyncMap) updateFile(protocol string, keys []string) {
	m.cfg.AppendToKey(protocol, keys)

	if err := m.cfg.WriteConfig(m.cfg.GetFileName()); err != nil {
		log.Error().Msg("error writing config: " + err.Error())
		return
	}
}

// StoreConfig sets the value for a key.
// It also updates the viper config file to keep new changes up to date.
// Overrides Store from sync.Map
func (m *MapSyncMap) StoreConfig(lName, key, protocol string, value interface{}) {
	if syncMap, _ := m.Load(lName); syncMap == nil {
		m.Store(lName, new(sync.Map))
	}
	syncMap, _ := m.Load(lName)
	syncMap.(*sync.Map).Store(key, value)

	// OUT is because we only handle the IPaddresses we had to protocol on at some point.
	// They are the one we need to manually reconnect if a problem happens.
	syncMap, _ = m.Load(lName)
	keys := Keys(syncMap.(*sync.Map), ALL) //OUT
	fmt.Println("(StoreConfig) keys : ", keys)
	if len(keys) != 0 {
		m.updateFile(protocol, keys)
	}
}

// DeleteConfig deletes the value for a key.
// It also updates the viper config file to keep new changes up to date by saving some data to handle before deleting the value.
// Overrides Delete from sync.Map
func (m *MapSyncMap) DeleteConfig(lName, connIpAddress string) {
	//fmt.Println("lname : ", lName, "connIpAddress : ", connIpAddress)

	//fmt.Println("MapSyncMap : ",m)

	syncMap, _ := m.Load(lName)
	conn, _ := syncMap.(*sync.Map).Load(connIpAddress)

	//fmt.Println("syncMap : ", syncMap.(*sync.Map)) //.Load(lName)
	//fmt.Println("conn : ", conn.(*ShosetConn))

	// OUT is because we only handle the IPaddresses we had to protocol on at some point.
	// They are the one we need to manually reconnect if a problem happens.
	ipAddresses := Keys(syncMap.(*sync.Map), OUT)

	//fmt.Println("LnamesByProtocol : ", conn.(*ShosetConn).GetShoset().LnamesByProtocol)

	conn.(*ShosetConn).GetShoset().LnamesByProtocol.Range(func(protocol, syncMap interface{}) bool {
		// LnamesByProtocol[protocol][lName]false || protocol != "bye"
		//fmt.Println("(func) syncMap : ", syncMap.(*sync.Map))
		if lnameInLnamesByProtocol, _ := syncMap.(*sync.Map).Load(lName); !lnameInLnamesByProtocol.(bool) || protocol != PROTOCOL_EXIT {
			return false
		}
		m.prepareUpdateFile(protocol.(string), ipAddresses)
		return true
	})

	syncMap.(*sync.Map).Delete(connIpAddress)
}

// Iterate calls *MapSync.Range(iter) sequentially for each key and *MapSync present in the map.
func (m *MapSyncMap) Iterate(iter func(string, interface{})) {
	for _, key := range m.Keys(ALL) {
		syncMap, _ := m.Load(key)
		if syncMap != nil {
			//fmt.Println("(Iterate) syncMap : ",syncMap.(*sync.Map))
			syncMap.(*sync.Map).Range(func(key, value interface{}) bool {
				iter(key.(string), value)
				return true
			})
		}
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

// Apppend a valye to a MapSyncMap : m[key1][key2]=value
func (m *MapSyncMap) AppendToKey(key1 string, key2 string, value interface{}) {
	if mapSync, ok := m.Load(key1); ok {
		mapSync.(*sync.Map).Store(key2, true)
		m.Store(key1, mapSync)
	} else {
		mapSync := new(sync.Map)
		mapSync.Store(key2, true)
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
