package concurentData

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type ConcurentSlice struct {
	sliceValues   []string
	contentChange chan bool
	m             sync.Mutex
}

const (
	TIMEOUT int = 5
)

func NewConcurentSlice() ConcurentSlice {
	return ConcurentSlice{contentChange: make(chan bool)}
}
func (cSlice *ConcurentSlice) AppendToConcurentSlice(data string) {
	cSlice.m.Lock()
	defer cSlice.m.Unlock()

	cSlice.sliceValues = append(cSlice.sliceValues, data)

	cSlice.sendChangeEvent()

	//fmt.Println("cSlice Append : ", cSlice, "data : ", data)
}

func (cSlice *ConcurentSlice) DeleteFromConcurentSlice(data string) {
	cSlice.m.Lock()
	defer cSlice.m.Unlock()
	//defer fmt.Println("cSlice Delete : ", cSlice, "data", data)

	for i, a := range cSlice.sliceValues {
		if a == data {
			cSlice.sliceValues[i] = cSlice.sliceValues[len(cSlice.sliceValues)-1]
			cSlice.sliceValues = cSlice.sliceValues[:len(cSlice.sliceValues)-1]

			cSlice.sendChangeEvent()

			return
		}
	}
	//fmt.Println("Failed to delete ", data)
}

func (cSlice *ConcurentSlice) sendChangeEvent() {
	select {
	case cSlice.contentChange <- true:
	default:
		fmt.Println("Nobody is waiting for contentChange")
	}
}

// Wait for the Slice to be empty
func (cSlice *ConcurentSlice) WaitForEmpty() error {
	for {
		cSlice.m.Lock()
		len := len(cSlice.sliceValues)
		cSlice.m.Unlock()
		if len != 0 {
			select {
			case <-cSlice.contentChange:
				fmt.Println("Received contentChange.")
				break
				//continue
			case <-time.After(time.Duration(TIMEOUT) * time.Second):
				return errors.New("the list is no empty (timeout)")
			}
		} else {
			return nil
		}
	}
}

func (cSlice *ConcurentSlice) String() string {
	// cSlice.m.Lock()
	// defer cSlice.m.Unlock()

	return fmt.Sprint(cSlice.sliceValues)
}
