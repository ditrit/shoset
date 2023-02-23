package concurentData

import (
	"errors"
	"fmt"
	"sync"
	"time"

	eventBus "github.com/ditrit/shoset/event_bus"
)

// This is a thread safe slice that allows you to wait for the next change or for it to be empty.
type ConcurentSlice struct {
	sliceValues []string
	eventBus    eventBus.EventBus // topics : change, empty (value not used)
	m           sync.RWMutex
}

func NewConcurentSlice() ConcurentSlice {
	return ConcurentSlice{eventBus: eventBus.NewEventBus()}
}

// Appends data to the slice and publishes the change event.
func (cSlice *ConcurentSlice) AppendToConcurentSlice(data string) {
	cSlice.m.Lock()
	defer cSlice.m.Unlock()

	cSlice.sliceValues = append(cSlice.sliceValues, data)
	cSlice.eventBus.Publish("change", true)
}

// Deletes data from the slice and publishes the change (and empty) event.
func (cSlice *ConcurentSlice) DeleteFromConcurentSlice(data string) {
	cSlice.m.Lock()
	defer cSlice.m.Unlock()

	for i, a := range cSlice.sliceValues {
		if a == data {
			// Deletes data from the slice in an efficient way (avoids moving remaining data).
			cSlice.sliceValues[i] = cSlice.sliceValues[len(cSlice.sliceValues)-1]
			cSlice.sliceValues = cSlice.sliceValues[:len(cSlice.sliceValues)-1]

			cSlice.eventBus.Publish("change", true)

			if len(cSlice.sliceValues) == 0 {
				cSlice.eventBus.Publish("empty", true)
			}
			return
		}
	}
}

// Waits for the Slice to be empty
func (cSlice *ConcurentSlice) WaitForEmpty(timeout int) error {	
	cSlice.m.RLock()
	if len(cSlice.sliceValues) != 0 {
		// Subscribes a channel to the empty topic :
		chEmpty := make(chan interface{})
		cSlice.eventBus.Subscribe("empty", chEmpty)
		defer cSlice.eventBus.UnSubscribe("empty", chEmpty)

		cSlice.m.RUnlock()
		select {
		case <-chEmpty:
			return nil
		case <-time.After(time.Duration(timeout) * time.Second):
			return errors.New("the list is no empty (timeout)")
		}
	}
	cSlice.m.RUnlock()
	return nil
}

// Wait for the Slice to change
func (cSlice *ConcurentSlice) WaitForChange(timeout int) error {
	cSlice.m.RLock()

	// Subscribes a channel to the change topic :
	chChange := make(chan interface{})
	cSlice.eventBus.Subscribe("change", chChange)
	defer cSlice.eventBus.UnSubscribe("change", chChange)

	cSlice.m.RUnlock()
	select {
	case <-chChange:
		return nil
	case <-time.After(time.Duration(timeout) * time.Second):
		return errors.New("the list did not change (timeout)")
	}
}

func (cSlice *ConcurentSlice) Contains(data string) bool {
	// contains range through a slice to search for a particular string
	for _, v := range cSlice.sliceValues {
		if v == data {
			return true
		}
	}
	return false
}

func (cSlice *ConcurentSlice) String() string {
	cSlice.m.RLock()
	defer cSlice.m.RUnlock()
	return fmt.Sprint(cSlice.sliceValues)
}
