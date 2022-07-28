package eventBus

import (
	"errors"
	"fmt"
	"sync"
)

// Allows to subscribe a channel to a topic and receive Events of that topic (the channel is only used to receive)

type DataChannel chan interface{}

// DataChannelSlice is a slice of DataChannels
type DataChannelSlice []DataChannel

// EventBus stores the information about subscribers interested for a particular topic
type EventBus struct {
	subscribers map[string]DataChannelSlice
	m           sync.Mutex
}

func NewEventBus() EventBus {
	return EventBus{
		subscribers: map[string]DataChannelSlice{},
	}
}

// Publishes some data on some topic, data is sent to every subscriber of the topic
func (eb *EventBus) Publish(topic string, data interface{}) {
	eb.m.Lock()
	defer eb.m.Unlock()

	//fmt.Println("Data : ", data, "topic : ", topic)
	if chans, found := eb.subscribers[topic]; found {
		// this is done because the slices refer to same array even though they are passed by value
		// thus we are creating a new slice with our elements thus preserve locking correctly.

		// Sends data to every channels subscribed to the topic
		//(goroutine to avoid blocking if the publisher and and the receiver are on the same goroutine)
		channels := append(DataChannelSlice{}, chans...)
		for _, ch := range channels {
			go func(data interface{}, Channel DataChannel) {
				defer func() {
					recover() // Avoids panicking when the channel was closed (by unsubscribing) before the send was completed
				}()
				//fmt.Println("Publishing to topic :", topic)
				Channel <- data
				// Add timeout
			}(data, ch)
		}
	} else {
		//fmt.Println("Nobody is subscribed to this topic : ", topic)
	}
}

func (eb *EventBus) Subscribe(topic string, ch DataChannel) {
	eb.m.Lock()
	defer eb.m.Unlock()

	//fmt.Println("Subscribing to topic : ", topic)

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

				//fmt.Println("UnSubscribing from topic : ", topic)

				close(ch)
				last := len(prev) - 1
				prev[i], prev[last] = prev[last], prev[i]
				prev = prev[:last]

				eb.subscribers[topic] = prev
				return nil
			}
		}
		return errors.New(" this chnnel  was not subscribed to this topic")
	}
	return errors.New(" topic not found")
}

func printDataEvent(ch string, data interface{}) {
	fmt.Printf("Channel: %s; DataEvent: %v\n", ch, data)
}
