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
		func(key string, val interface{}) error {
			fmt.Printf(" - %s: %d\n", key, val.(int))
			return nil
		},
	)
	m.Set("a", 454433)
	fmt.Printf("After update : ")
	m.Iterate(
		func(key string, val interface{}) error {
			fmt.Printf(" - %s: %d\n", key, val.(int))
			return nil
		},
	)
	m.Delete("b")
	fmt.Printf("After delete : ")
	m.Iterate(
		func(key string, val interface{}) error {
			fmt.Printf(" - %s: %d\n", key, val.(int))
			return nil
		},
	)
}

// TestFold : test MapSafe folding using closure
func TestFold(t *testing.T) {
	m := shoset.NewMapSafe()
	m.Set("a", 23).Set("b", 43).Set("c", 11)
	str := ""
	m.Iterate(
		func(key string, val interface{}) error {
			str = fmt.Sprintf("%s, %d", str, val.(int))
			fmt.Printf(" - %s: %d\n", key, val.(int))
			return nil
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
		func(key string, val interface{}) error {
			time.Sleep(time.Millisecond * time.Duration(10))
			fmt.Printf("%s, %d\n", key, val.(int))
			return nil
		})
	time.Sleep(time.Millisecond * time.Duration(20))
	fmt.Printf("after Iterate\n")
	m.Set("a", 101)
	m.Set("b", 102)
	fmt.Printf("after Set\n")
	m.Iterate(
		func(key string, val interface{}) error {
			fmt.Printf("%s, %d\n", key, val.(int))
			return nil
		})

}
