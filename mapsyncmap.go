package shoset

import (
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

// StoreConfig sets the conn for a address.
// It also updates the viper config file to keep new changes up to date.
// Overrides Store from sync.Map
func (m *MapSyncMap) StoreConfig(lName, address, protocol string, conn interface{}) error {
	var err error = nil
	if syncMap, _ := m.Load(lName); syncMap == nil {
		m.Store(lName, new(sync.Map))
	}
	syncMap, _ := m.Load(lName)
	syncMap.(*sync.Map).Store(address, conn)

	// OUT is because we only handle the IPaddresses we had to protocol on at some point.
	// They are the one we need to manually reconnect if a problem happens.
	syncMap, _ = m.Load(lName)
	addresses := Keys(syncMap.(*sync.Map), OUT)
	if len(addresses) != 0 {
		err = m.updateFile(protocol, addresses)
	}
	return err
}

// DeleteConfig deletes the value for a key.
// It also updates the viper config file to keep new changes up to date by saving some data to handle before deleting the value.
// Overrides Delete from sync.Map
func (m *MapSyncMap) DeleteConfig(lName, connIpAddress string) {
	syncMap, _ := m.Load(lName)
	conn, _ := syncMap.(*sync.Map).Load(connIpAddress)
	if conn != nil {
		syncMap.(*sync.Map).Delete(connIpAddress)
	}

	protocols := conn.(*ShosetConn).GetShoset().LnamesByProtocol.Keys(ALL)
	for _, protocol := range protocols {
		if storedIpAddresses := m.cfg.viper.Get(protocol); storedIpAddresses != nil {
			var newIpAddresses []string
			for _, storedIpAddress := range storedIpAddresses.([]string) {
				if storedIpAddress != connIpAddress {
					newIpAddresses = append(newIpAddresses, storedIpAddress)
				}
			}
			m.updateFile(protocol, newIpAddresses)
		}
	}
}

// updateFile updates config after being set
func (m *MapSyncMap) updateFile(protocol string, addresses []string) error {
	m.cfg.Set(protocol, addresses)

	if err := m.cfg.WriteConfig(m.cfg.GetFileName()); err != nil {
		log.Error().Msg("error writing config: " + err.Error())
		return err
	}
	return nil
}

// Iterate calls *MapSync.Range(iter) sequentially for each key and *MapSync present in the map.
func (m *MapSyncMap) Iterate(iter func(string, interface{})) {
	for _, key := range m.Keys(ALL) {
		syncMap, _ := m.Load(key)
		if syncMap != nil {
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
