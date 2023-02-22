package eventBus

// Subscribe a channel to a topic and receive Events of that topic (the channel is only used to receive)
/* CAUTION :
-Channels are not resuable, they are closed when unsubscribing.
-Sending more events than are read can create massive memory leaks.
*/

// Credit : original code from : https://levelup.gitconnected.com/lets-write-a-simple-event-bus-in-go-79b9480d8997 (modified)

import (
	"errors"
	"sync"
)

type DataChannel chan interface{}

type DataChannelSlice []DataChannel

// EventBus stores the information about subscribers interested in a particular topic
type EventBus struct {
	subscribers map[string]DataChannelSlice //map[topic] (list of channels subscribed to the topic)
	m           sync.RWMutex
}

func NewEventBus() EventBus {
	return EventBus{
		subscribers: map[string]DataChannelSlice{},
	}
}

// Publishes some data on some topic, data is sent to every subscriber of the topic
func (eb *EventBus) Publish(topic string, data interface{}) {
	eb.m.RLock()
	defer eb.m.RUnlock()

	if chans, found := eb.subscribers[topic]; found {
		// this is done because the slices refer to same array even though they are passed by value
		// thus we are creating a new slice with our elements thus preserve locking correctly.

		// Sends data to every channels subscribed to the topic
		//(goroutine to avoid waiting for the previous subscriber to read the event to send the next and to avoid deadlock if the publisher and and the receiver are on the same goroutine)
		channels := append(DataChannelSlice{}, chans...)
		for _, ch := range channels {
			go func(data interface{}, Channel DataChannel) {
				/*defer func() { // Avoids panicking when the channel was closed (by unsubscribing) before the send was completed
					recover()
				}()*/
				Channel <- data
			}(data, ch)
		}
	}
}

func (eb *EventBus) Subscribe(topic string, ch DataChannel) {
	eb.m.Lock()
	defer eb.m.Unlock()

	if prev, found := eb.subscribers[topic]; found {
		eb.subscribers[topic] = append(prev, ch)
	} else {
		eb.subscribers[topic] = append([]DataChannel{}, ch)
	}
}

func (eb *EventBus) UnSubscribe(topic string, ch DataChannel) error {
	eb.m.Lock()
	defer eb.m.Unlock()

	// Deletes the channel from the slice
	if prev, found := eb.subscribers[topic]; found {
		for i, a := range eb.subscribers[topic] {
			if a == ch {
				//close(ch)

				// Deletes channel from subscribers to the topic in an efficient way (avoids moving remaining data).
				last := len(prev) - 1
				prev[i], prev[last] = prev[last], prev[i]
				prev = prev[:last]

				eb.subscribers[topic] = prev
				return nil
			}
		}
		return errors.New(" this channel was not subscribed to this topic : " + topic)
	}
	return errors.New(" topic not found : " + topic)
}
