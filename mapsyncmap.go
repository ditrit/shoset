package shoset

import (
	"sync"

	"github.com/rs/zerolog/log"
)

// MapSyncMap : safe for concurrent use by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
type MapSyncMap struct {
	smap sync.Map // map[string]*sync.Map
	cfg      *Config
}

// SetConfig sets the config with new one.
func (mapSyncMap *MapSyncMap) SetConfig(cfg *Config) {
	mapSyncMap.cfg = cfg
}

// updateFile updates config after being set
func (mapSyncMap *MapSyncMap) updateFile(protocolType, fileName string, keys []string) {
	mapSyncMap.cfg.Set(protocolType, keys)

	if err := mapSyncMap.cfg.WriteConfig(fileName); err != nil {
		log.Error().Msg("error writing config: " + err.Error())
		return
	}
}

// Store sets the value for a key.
// It also updates the viper config file to keep new changes up to date.
func (mapSyncMap *MapSyncMap) Store(lName, key, protocolType, shosetType, fileName string, value interface{}) {
	mapSync := new(sync.Map)
	mapSync.Store(lName, true)
	value.(*ShosetConn).ch.LnamesByProtocol.smap.Store(protocolType, mapSync)
	value.(*ShosetConn).ch.LnamesByType.smap.Store(shosetType, mapSync)

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
		mapSyncMap.updateFile(protocolType, fileName, keys)
	}
}

// Delete deletes the value for a key.
// It also updates the viper config file to keep new changes up to date.
func (mapSyncMap *MapSyncMap) Delete(lname, key, fileName string) {
	var lNamesByProtocol *MapSyncMap
	var keys []string

	if mapSync, _ := mapSyncMap.smap.Load(lname); mapSync != nil {
		if conn, _ := mapSync.(*sync.Map).Load(key); conn != nil {
			lNamesByProtocol = conn.(*ShosetConn).ch.LnamesByProtocol
		}

		// OUT is because we only handle the IPaddresses we had to protocol on at some point.
		// They are the one we need to manually reconnect if a problem happens.
		keys = Keys(mapSync.(*sync.Map), OUT)

		mapSync.(*sync.Map).Delete(key)
	}

	if lNamesByProtocol == nil {
		return
	}
	var protocolTypes []string
	lNamesByProtocol.smap.Range(func(key, value interface{}) bool {
		func(protocol string, lNames interface{}) {
			// if lname not in lNames
			lnameInlnames, _ := lNames.(*sync.Map).Load(lname)
			if lnameInlnames != nil {
				if !lnameInlnames.(bool) || protocol != PROTOCOL_EXIT {
					return
				}
			}
	
			if contains(mapSyncMap.cfg.GetSlice(PROTOCOL_JOIN), key.(string)) {
				protocolTypes = append(protocolTypes, PROTOCOL_JOIN)
			}
			if contains(mapSyncMap.cfg.GetSlice(PROTOCOL_LINK), key.(string)) {
				protocolTypes = append(protocolTypes, PROTOCOL_LINK)
			}
	
			for _, protocolType := range protocolTypes {
				mapSyncMap.updateFile(protocolType, fileName, keys)
			}
		}(key.(string), value)
		return true
	})
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
	return keys
}