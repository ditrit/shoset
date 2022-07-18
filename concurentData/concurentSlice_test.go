package concurentData

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Ajouter des données X
// Supprimer des données X
// Oppération concourantes X

func Test_concurentSlice(t *testing.T) {
	cSlice := NewConcurentSlice()

	number := 10

	for i := 0; i < number; i++ {
		cSlice.AppendToConcurentSlice("Test " + fmt.Sprint(i))
	}

	if cSlice.String() != "[Test 0 Test 1 Test 2 Test 3 Test 4 Test 5 Test 6 Test 7 Test 8 Test 9]" {
		t.Errorf("Wrong content after adding.")
	}

	fmt.Println(cSlice.String())

	for i := 0; i < number-5; i++ {
		cSlice.DeleteFromConcurentSlice("Test " + fmt.Sprint(i))
	}

	if cSlice.String() != "[Test 9 Test 8 Test 7 Test 6 Test 5]" {
		t.Errorf("Wrong content after Clearing.")
	}

	fmt.Println(cSlice.String())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < number-5; i++ {
			cSlice.AppendToConcurentSlice("Test " + fmt.Sprint(i))
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 5; i < number; i++ {
			cSlice.DeleteFromConcurentSlice("Test " + fmt.Sprint(i))
		}
	}()

	wg.Wait()

	go func() {
		for i := 0; i < number-5; i++ {
			cSlice.DeleteFromConcurentSlice("Test " + fmt.Sprint(i))
			fmt.Println(cSlice.String())
			time.Sleep(10 * time.Millisecond)
		}
	}()

	cSlice.WaitForEmpty()

	if cSlice.String() != "[]" {
		t.Errorf("Wrong content after Clearing.")
	}

	fmt.Println(cSlice.String())

	cSlice.AppendToConcurentSlice("TEST")

	err := cSlice.WaitForEmpty()

	if err==nil{
		t.Errorf("Timeout error failed.")
	}
}
