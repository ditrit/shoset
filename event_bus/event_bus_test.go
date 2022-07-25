package eventBus

import (
	"fmt"
	"testing"
	"time"
)

func TestFeed_simple(t *testing.T) {
	var eb = NewEventBus()

	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	ch3 := make(chan interface{})

	eb.Subscribe("topic1", ch1)
	eb.Subscribe("topic2", ch2)
	eb.Subscribe("topic3", ch3)

	go eb.Publish("topic1", "Hi topic 1")
	go eb.Publish("topic2", "Welcome to topic 2")
	go eb.Publish("topic3", "Welcome to topic 3")

	for i := 0; i < 3; i++ {
		select {
		case d := <-ch1:
			printDataEvent("ch1", d)
		case d := <-ch2:
			printDataEvent("ch2", d)
		case d := <-ch3:
			printDataEvent("ch3", d)
		}
	}
}

func TestFeed_UnSubscribe(t *testing.T) {
	var eb = NewEventBus()

	ch1 := make(chan interface{})

	eb.Subscribe("topic1", ch1)

	fmt.Println(eb.subscribers["topic1"])

	eb.Publish("topic1", "Hi topic 1")
	d, ok := <-ch1
	if !ok {
		t.Errorf("Channel closed when it shouldn't.")
	}
	printDataEvent("ch1", d)

	eb.UnSubscribe("topic1", ch1)

	fmt.Println(eb.subscribers["topic1"])
}

func TestFeed_ManyMessages(t *testing.T) {
	var eb = NewEventBus()

	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	ch3 := make(chan interface{})

	eb.Subscribe("topic1", ch1)
	eb.Subscribe("topic2", ch2)
	eb.Subscribe("topic3", ch3)

	// Sender
	go func() {
		for {
			eb.Publish("topic1", "topic 1")
			eb.Publish("topic2", "topic 2")
			eb.Publish("topic3", "topic 3")
			//time.Sleep(5 * time.Millisecond)
		}
	}()

	timer := time.NewTimer(2 * time.Second)

	received :=0

receive:
	for {
		select {
		case <-ch1:
		case <-ch2:
		case <-ch3:
		case <-timer.C:
			break receive
		}
		received++

	}

	fmt.Println("Number of message received : ",received)
}
