package shoset

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// MapSyncMap : safe for concurrent use by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
type MapSyncMap struct {
	smap sync.Map // map[string]*sync.Map
	cfg  *Config
}

// SetConfig sets the config with new one.
func (mapSyncMap *MapSyncMap) GetConfig() *Config {
	return mapSyncMap.cfg
}

// SetConfig sets the config with new one.
func (mapSyncMap *MapSyncMap) SetConfig(cfg *Config) {
	mapSyncMap.cfg = cfg
}

func (mapSyncMap *MapSyncMap) prepareUpdateFile(protocol string, ipAddresses []string) {
	var protocols []string

	if contains(mapSyncMap.cfg.GetSlice(PROTOCOL_JOIN), protocol) {
		protocols = append(protocols, PROTOCOL_JOIN)
	}
	if contains(mapSyncMap.cfg.GetSlice(PROTOCOL_LINK), protocol) {
		protocols = append(protocols, PROTOCOL_LINK)
	}

	for _, protocol := range protocols {
		mapSyncMap.updateFile(protocol, ipAddresses)
	}
}

// updateFile updates config after being set
func (mapSyncMap *MapSyncMap) updateFile(protocol string, keys []string) {
	mapSyncMap.cfg.Set(protocol, keys)

	if err := mapSyncMap.cfg.WriteConfig(mapSyncMap.cfg.GetFileName()); err != nil {
		log.Error().Msg("error writing config: " + err.Error())
		return
	}
}

// Store sets the value for a key.
// It also updates the viper config file to keep new changes up to date.
// Overrides Store from sync.Map
func (mapSyncMap *MapSyncMap) Store(lName, key, protocol string, value interface{}) {
	if mapSync, _ := mapSyncMap.smap.Load(lName); mapSync == nil {
		mapSyncMap.smap.Store(lName, new(sync.Map))
	}
	newMapSync, _ := mapSyncMap.smap.Load(lName)
	newMapSync.(*sync.Map).Store(key, value)

	// OUT is because we only handle the IPaddresses we had to protocol on at some point.
	// They are the one we need to manually reconnect if a problem happens.
	newMapSync, _ = mapSyncMap.smap.Load(lName)
	keys := Keys(newMapSync.(*sync.Map), OUT)
	if len(keys) != 0 {
		mapSyncMap.updateFile(protocol, keys)
	}
}

// Delete deletes the value for a key.
// It also updates the viper config file to keep new changes up to date by saving some data to handle before deleting the value.
// Overrides Delete from sync.Map
func (mapSyncMap *MapSyncMap) Delete(lName, connIpAddress string) {
	mapSync, _ := mapSyncMap.smap.Load(lName)
	conn, _ := mapSync.(*sync.Map).Load(connIpAddress)

	// OUT is because we only handle the IPaddresses we had to protocol on at some point.
	// They are the one we need to manually reconnect if a problem happens.
	ipAddresses := Keys(mapSync.(*sync.Map), OUT)

	conn.(*ShosetConn).GetShoset().LnamesByProtocol.smap.Range(func(protocol, mapSync interface{}) bool {
		// LnamesByProtocol[protocol][lName]false || protocol != "bye"
		if lnameInLnamesByProtocol, _ := mapSync.(*sync.Map).Load(lName); !lnameInLnamesByProtocol.(bool) || protocol != PROTOCOL_EXIT {
			return false
		}
		mapSyncMap.prepareUpdateFile(protocol.(string), ipAddresses)
		return true
	})

	mapSync.(*sync.Map).Delete(connIpAddress)
}

// Iterate calls *MapSync.Range(iter) sequentially for each key and *MapSync present in the map.
func (mapSyncMap *MapSyncMap) Iterate(iter func(string, interface{})) {
	for _, key := range mapSyncMap.Keys(ALL) {
		mapSync, _ := mapSyncMap.smap.Load(key)
		if mapSync != nil {
			mapSync.(*sync.Map).Range(func(key, value interface{}) bool {
				iter(key.(string), value)
				return true
			})
		}
	}
}

// Keys returns a []string corresponding to the keys from the map[string]*sync.Map object.
// sType set the specific keys depending on the desired shoset type.
func (mapSyncMap *MapSyncMap) Keys(sType string) []string {
	var keys []string
	if sType == ALL {
		// all keys whatever the type from the ShosetConn
		mapSyncMap.smap.Range(func(key, _ interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
	} else {
		// keys with specific ShosetConn type
		mapSyncMap.smap.Range(func(key, _ interface{}) bool {
			if key == sType {
				keys = append(keys, key.(string))
			}
			return true
		})
	}

	return removeDuplicateStr(keys)
}
