package concurentData

import (
	"fmt"
	"sync"
)

type ConcurentSlice struct {
	sliceValues   []string
	contentChange chan bool
	m             sync.Mutex
}

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
func (cSlice *ConcurentSlice) WaitForEmpty() {
	for {
		cSlice.m.Lock()
		len := len(cSlice.sliceValues)
		cSlice.m.Unlock()
		if  len != 0 {
			<-cSlice.contentChange
			fmt.Println("Received contentChange.")
			continue
		} else {
			return
		}
	}
}

func (cSlice *ConcurentSlice) String() string {
	// cSlice.m.Lock()
	// defer cSlice.m.Unlock()

	return fmt.Sprint(cSlice.sliceValues)
}
