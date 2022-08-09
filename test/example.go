package main // tests run in the main package

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
	utilsForTest "github.com/ditrit/shoset/test/utils_for_test"
)

// Simplest example :
func simpleExample() {
	cl1 := shoset.NewShoset("A", "TypeOfA")
	cl1.InitPKI("localhost:8001") // Is the CA of the network

	cl2 := shoset.NewShoset("B", "TypeOfB")
	cl2.Protocol("localhost:8002", "localhost:8001", "link") // we link it to our first socket

	cl1.WaitForProtocols(5) // Wait for cl1 to be ready
	cl2.WaitForProtocols(5)

	//Sender :
	event := msg.NewEventClassic("test_topic", "test_event", "test_payload")
	cl2.Send(event)

	//Receiver :
	event_rc := cl1.Wait("evt", map[string]string{"topic": "test_topic", "event": "test_event"}, 5, nil)
	fmt.Println("event received : ", event_rc)
	shoset.Log("event received (Payload) : " + event_rc.GetPayload())
}

// Send an event every second forever :
func testEventContinuousSend() {
	cl1 := shoset.NewShoset("A", "cl")
	cl1.InitPKI("localhost:8001") // Is the CA of the network

	cl2 := shoset.NewShoset("B", "cl")
	cl2.Protocol("localhost:8002", "localhost:8001", "link") // we link it to our first socket

	cl1.WaitForProtocols(5) // Wait for cl1 to be ready
	cl2.WaitForProtocols(5)

	//Sender :
	go func() {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			event := msg.NewEventClassic("test_topic", "test_event", "test_payload"+fmt.Sprint(i))
			cl2.Send(event)
			//fmt.Println("event (sender) : ", event)
			i++
		}
	}()

	//Receiver :
	iterator := msg.NewIterator(cl1.Queue["evt"])
	go func() {
		for {
			event_rc := cl1.Wait("evt", map[string]string{"topic": "test_topic", "event": "test_event"}, 5, iterator)
			fmt.Println("event received : ", event_rc)
			shoset.Log("event received (Payload) : " + event_rc.GetPayload())
		}
	}()

	// Do someting else while it is sending and receiving messages.

	select {} //Never return
}

// Forwarding : Send a message every second from C to A forever.
func testSimpleForwarding() {
	cl1 := shoset.NewShoset("A", "cl")
	cl1.InitPKI("localhost:8001") // Is the CA of the network

	cl2 := shoset.NewShoset("B", "cl")
	cl2.Protocol("localhost:8002", "localhost:8001", "link") // we link it to our first socket

	cl3 := shoset.NewShoset("C", "cl")
	cl3.Protocol("localhost:8003", "localhost:8002", "link")

	cl1.WaitForProtocols(5) // Wait for cl1 to be ready
	cl2.WaitForProtocols(5)
	cl3.WaitForProtocols(5)

	//Sender :
	go func() {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			event := msg.NewSimpleMessage("A", "test_payload"+fmt.Sprint(i))
			cl3.Send(event)
			i++
		}
	}()

	//Receiver :
	iterator := msg.NewIterator(cl1.Queue["simpleMessage"])
	go func() {
		for {
			event_rc := cl1.Wait("simpleMessage", map[string]string{}, 5, iterator)
			fmt.Println("message received : ", event_rc)
			if event_rc != nil {
				shoset.Log("message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()

	// Do someting else while it is sending and receiving messages.

	select {} //Never return
}

// Forwarding using topology : Same as before but using the topology sytem to symplify setup.
func testForwardingTopology() {
	tt := utilsForTest.Circle // Choose the network topology for the test

	s := []*shoset.Shoset{} // Create the empty list of shosets

	s = utilsForTest.CreateManyShosets(tt, s, false) // Populate the list with the shosets as specified in the selected topology and estavlish connection among them

	utilsForTest.WaitForManyShosets(s) // Wait for every shosets in the list to be ready

	utilsForTest.PrintManyShosets(s) // Display the info of every shosets in the list

	destination := s[len(s)-1]
	sender := s[0]

	//Sender :
	go func() {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			event := msg.NewSimpleMessage(destination.GetLogicalName(), "test_payload"+fmt.Sprint(i))
			sender.Send(event)
			i++
		}
	}()

	//Receiver :
	iterator := msg.NewIterator(destination.Queue["simpleMessage"])
	go func() {
		for {
			event_rc := destination.Wait("simpleMessage", map[string]string{}, 5, iterator)
			fmt.Println("message received : ", event_rc)
			if event_rc != nil {
				shoset.Log("message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()

	// Do someting else while it is sending and receiving messages.

	time.Sleep(10*time.Second)

	//select {} //Never return
}
