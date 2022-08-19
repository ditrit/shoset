package concurentData

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func Test_concurentSlice(t *testing.T) {
	cSlice := NewConcurentSlice()

	number := 10 //Changing this requires updating the strings the result is compared to.

	go func() {
		for {
			cSlice.WaitForChange(5)
			fmt.Println("Change received.")
		}
	}()

	time.Sleep(10 * time.Millisecond)

	for i := 0; i < number; i++ {
		cSlice.AppendToConcurentSlice("Test " + fmt.Sprint(i))
	}
	if !cSlice.Contains("Test 0") {
		t.Errorf("Wrong content after appending.")
	}
	if cSlice.String() != "[Test 0 Test 1 Test 2 Test 3 Test 4 Test 5 Test 6 Test 7 Test 8 Test 9]" {
		t.Errorf("Wrong content after appending.")
	}
	fmt.Println(cSlice.String())

	for i := 0; i < number-(number/2); i++ {
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
		for i := 0; i < number-(number/2); i++ {
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
		for i := 0; i < number-(number/2); i++ {
			cSlice.DeleteFromConcurentSlice("Test " + fmt.Sprint(i))
			fmt.Println(cSlice.String())
			time.Sleep(10 * time.Millisecond)
		}
	}()

	cSlice.WaitForEmpty(5)

	if cSlice.String() != "[]" {
		t.Errorf("Wrong content after Clearing.")
	}

	fmt.Println(cSlice.String())

	cSlice.AppendToConcurentSlice("TEST")

	err := cSlice.WaitForEmpty(5)

	if err == nil {
		t.Errorf("Timeout error failed.")
	}
}
