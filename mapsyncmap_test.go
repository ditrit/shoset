package shoset

import (
	"strconv"
	"sync"
	"testing"
)

var shoset *Shoset = NewShoset("cl", "cl")

func TestStoreConfig(t *testing.T) {
	shoset.InitPKI("127.0.0.1:9001")

	direction := []string{IN, OUT}

	for i := 2; i < 6; i++ {
		conn, err := NewShosetConn(shoset, "127.0.0.1:900"+strconv.Itoa(i), direction[i%len(direction)])
		if err != nil {
			t.Errorf("StoreConfig didn't work, conn is nil")
		}
		err = shoset.ConnsByLname.StoreConfig("test", "127.0.0.1:900"+strconv.Itoa(i), "test_protocol", conn)
		if err != nil {
			t.Errorf("StoreConfig didn't work")
		}
		mapSync := new(sync.Map)
		mapSync.Store("test", true)
		shoset.LnamesByProtocol.Store("test_protocol", mapSync)
	}
}

func TestKeys(t *testing.T) {
	TestStoreConfig(t)

	keys := shoset.ConnsByLname.Keys(ALL)
	if keys[0] != "test" {
		t.Errorf("Keys didn't work")
	}
}
