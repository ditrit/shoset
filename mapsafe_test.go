package shoset_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ditrit/shoset"
)

// TestMapSafeCRUD : test MapSafe crud functions
func TestMapSafeCRUD(t *testing.T) {
	m := shoset.NewMapSafe()
	m.Set("a", 23).Set("b", 43).Set("c", 11)
	fmt.Printf("Initial state : ")
	m.Iterate(
		func(key string, val interface{}) {
			fmt.Printf(" - %s: %d\n", key, val.(int))
		},
	)
	m.Set("a", 454433)
	fmt.Printf("After update : ")
	m.Iterate(
		func(key string, val interface{}) {
			fmt.Printf(" - %s: %d\n", key, val.(int))
		},
	)
	m.Delete("b")
	fmt.Printf("After delete : ")
	m.Iterate(
		func(key string, val interface{}) {
			fmt.Printf(" - %s: %d\n", key, val.(int))
		},
	)
}

// TestFold : test MapSafe folding using closure
func TestFold(t *testing.T) {
	m := shoset.NewMapSafe()
	m.Set("a", 23).Set("b", 43).Set("c", 11)
	str := ""
	m.Iterate(
		func(key string, val interface{}) {
			str = fmt.Sprintf("%s, %d", str, val.(int))
			fmt.Printf(" - %s: %d\n", key, val.(int))
		})
	fmt.Printf("strValue : %s\n", str)
}

// TestConcurrency
func TestConcurrency(t *testing.T) {
	m := shoset.NewMapSafe()
	for i := 0; i < 10; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	fmt.Printf("test Concurrency\n")
	go m.Iterate(
		func(key string, val interface{}) {
			time.Sleep(time.Millisecond * time.Duration(10))
			fmt.Printf("%s, %d\n", key, val.(int))
		})
	time.Sleep(time.Millisecond * time.Duration(20))
	fmt.Printf("after Iterate\n")
	m.Set("a", 101)
	m.Set("b", 102)
	fmt.Printf("after Set\n")
	m.Iterate(
		func(key string, val interface{}) {
			fmt.Printf("%s, %d\n", key, val.(int))
		})

}
