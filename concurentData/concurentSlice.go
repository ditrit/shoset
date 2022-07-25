package concurentData

import (
	"errors"
	"fmt"
	"sync"
	"time"

	eventBus "github.com/ditrit/shoset/event_bus"
)

type ConcurentSlice struct {
	sliceValues []string
	eventBus    eventBus.EventBus // topics : change, empty
	m           sync.Mutex
}

const (
	TIMEOUT int = 5
)

func NewConcurentSlice() ConcurentSlice {
	return ConcurentSlice{eventBus: eventBus.NewEventBus()}
}
func (cSlice *ConcurentSlice) AppendToConcurentSlice(data string) {
	cSlice.m.Lock()
	defer cSlice.m.Unlock()

	cSlice.sliceValues = append(cSlice.sliceValues, data)

	fmt.Println("Sending change")
	cSlice.eventBus.Publish("change", true)

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

			fmt.Println("Sending change")
			cSlice.eventBus.Publish("change", true)

			if len(cSlice.sliceValues) == 0 {
				fmt.Println("Sending empty")
				cSlice.eventBus.Publish("empty", true)
			}

			return
		}
	}
	//fmt.Println("Failed to delete ", data)
}

// Wait for the Slice to be empty
func (cSlice *ConcurentSlice) WaitForEmpty() error {
	// Subscribe a channel to the empty topic :
	cSlice.m.Lock()

	chEmpty := make(chan interface{})
	cSlice.eventBus.Subscribe("empty", chEmpty)

	defer cSlice.eventBus.UnSubscribe("empty", chEmpty)

	if len(cSlice.sliceValues) != 0 {
		cSlice.m.Unlock()
		select {
		case <-chEmpty:
			fmt.Println("Received Empty")
			//cSlice.eventBus.UnSubscribe("empty", chEmpty)
			return nil

		case <-time.After(time.Duration(TIMEOUT) * time.Second):
			//cSlice.eventBus.UnSubscribe("empty", chEmpty)
			return errors.New("the list is no empty (timeout)")
		}
	}
	cSlice.m.Unlock()
	return nil
}

// Wait for the Slice to change
func (cSlice *ConcurentSlice) WaitForChange() error {
	// Subscribe a channel to the empty topic :
	cSlice.m.Lock()

	chChange := make(chan interface{})
	cSlice.eventBus.Subscribe("change", chChange)

	defer cSlice.eventBus.UnSubscribe("change", chChange)

	cSlice.m.Unlock()

	//fmt.Println("Subscribed to change")
	select {
	case <-chChange:
		fmt.Println("Received change")
		//cSlice.eventBus.UnSubscribe("change", chChange)
		return nil

	case <-time.After(time.Duration(TIMEOUT) * time.Second):
		//cSlice.eventBus.UnSubscribe("change", chChange)
		return errors.New("the list did not change (timeout)")
	}
}

func (cSlice *ConcurentSlice) String() string {
	return fmt.Sprint(cSlice.sliceValues)
}
