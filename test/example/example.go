package example

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
	utilsForTest "github.com/ditrit/shoset/test/utils_for_test"
)

// Simplest example :
func SimpleExample() {
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
	if event_rc != nil {
		shoset.Log("event received (Payload) : " + event_rc.GetPayload())
	}
}

// Send an event every second forever :
func TestEventContinuousSend() {
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
			if event_rc != nil {
				shoset.Log("event received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()

	// Do someting else while it is sending and receiving messages.

	select {} //Never return
}

// Forwarding : Send a message every second from C to A forever.
func TestSimpleForwarding() {
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
func TestForwardingTopology() {
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
			message := msg.NewSimpleMessage(destination.GetLogicalName(), "test_payload"+fmt.Sprint(i))
			sender.Send(message)
			i++
			if i >= 10 {
				break
			}
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

	time.Sleep(5 * time.Second) // Wait for the routing to be established
	utilsForTest.PrintManyShosets(s)
	time.Sleep(10 * time.Second) // Wait for the end of the test

	//select {} //Never return
}

func TestMutltipleShosets() {
	fmt.Println("TestMultipleShosets")
	tt := utilsForTest.Line2With3Shosets // Choose the network topology for the test

	s := []*shoset.Shoset{} // Create the empty list of shosets

	s = utilsForTest.CreateManyShosets(tt, s, false) // Populate the list with the shosets as specified in the selected topology and estavlish connection among them

	utilsForTest.WaitForManyShosets(s) // Wait for every shosets in the list to be ready

	time.Sleep(3 * time.Second)

	utilsForTest.PrintManyShosets(s) // Display the info of every shosets in the list

	time.Sleep(20 * time.Second)

	//select {} //Never return
}

func Test3Shosets() {

	tt := utilsForTest.Line2With2Shosets // Choose the network topology for the test

	s := []*shoset.Shoset{} // Create the empty list of shosets

	s = utilsForTest.CreateManyShosets(tt, s, false) // Populate the list with the shosets as specified in the selected topology and estavlish connection among them

	utilsForTest.WaitForManyShosets(s) // Wait for every shosets in the list to be ready

	time.Sleep(3 * time.Second)

	utilsForTest.PrintManyShosets(s) // Display the info of every shosets in the list

	//Sender :
	go func() {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			event := msg.NewEventClassic("test_topic", "test_event", "test_payload"+fmt.Sprint(i))
			s[0].Send(event)
			fmt.Println("A send event to B1 : ", "test_payload"+fmt.Sprint(i))
			i++
		}
	}()

	//Receiver :
	iterator := msg.NewIterator(s[1].Queue["evt"])
	go func() {
		for {
			event_rc := s[1].Wait("evt", map[string]string{"topic": "test_topic", "event": "test_event"}, 5, iterator)
			if event_rc != nil {
				fmt.Println("B1 message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()

	//Receiver :
	iterator2 := msg.NewIterator(s[2].Queue["evt"])
	go func() {
		for {
			event_rc := s[2].Wait("evt", map[string]string{"topic": "test_topic", "event": "test_event"}, 5, iterator2)
			if event_rc != nil {
				fmt.Println("B2 message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()

	time.Sleep(20 * time.Second)
}

// test with A-B-B
func Test3ShosetsCommand() {
	fmt.Println("Test3ShosetsCommand : \n A-B1-B2 \n A send 'command' to B1 \n Normally the command is transmitted to B2")
	tt := utilsForTest.Line2With2Shosets // Choose the network topology for the test

	s := []*shoset.Shoset{} // Create the empty list of shosets

	s = utilsForTest.CreateManyShosets(tt, s, false) // Populate the list with the shosets as specified in the selected topology and estavlish connection among them

	utilsForTest.WaitForManyShosets(s) // Wait for every shosets in the list to be ready

	//utilsForTest.PrintManyShosets(s) // Display the info of every shosets in the list

	//destination := s[1] // B far from B
	sender := s[0] // A pki

	//Sender :
	go func() {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			message := msg.NewCommand("B", "test_command", "test_payload"+fmt.Sprint(i))
			fmt.Println("A send command to B1 : ", "test_payload"+fmt.Sprint(i))
			sender.Send(*message)
			i++
			if i >= 10 {
				break
			}
		}
	}()

	//Receiver :
	iterator := msg.NewIterator(s[1].Queue["cmd"])
	go func() {
		for {
			event_rc := s[1].Wait("cmd", map[string]string{"name": "test_command"}, 5, iterator)
			if event_rc != nil {
				shoset.Log("B1 message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()
	iterator2 := msg.NewIterator(s[2].Queue["cmd"])
	rcvd := 0
	go func() {
		for {
			event_rc2 := s[2].Wait("cmd", map[string]string{"name": "test_command"}, 5, iterator2)
			if event_rc2 != nil {
				shoset.Log("B2 message received (Payload) : " + event_rc2.GetPayload())
				rcvd += 1
			}
		}
	}()

	// Do someting else while it is sending and receiving messages.

	time.Sleep(3 * time.Second) // Wait for the routing to be established

	time.Sleep(5 * time.Second) // Wait for the end of the test
	fmt.Println(rcvd)
	utilsForTest.PrintManyShosets(s) // Display the info of every shosets in the list

	//select {} //Never return
}

// test with A-B-B
func Test2ShosetsCommand() {
	fmt.Println("Test2ShosetsCommand : \n A-B \n A send 'command' to B")
	tt := utilsForTest.Line2 // Choose the network topology for the test

	s := []*shoset.Shoset{} // Create the empty list of shosets

	s = utilsForTest.CreateManyShosets(tt, s, false) // Populate the list with the shosets as specified in the selected topology and estavlish connection among them

	utilsForTest.WaitForManyShosets(s) // Wait for every shosets in the list to be ready

	//utilsForTest.PrintManyShosets(s) // Display the info of every shosets in the list

	destination := s[1] // B
	sender := s[0]      // A pki

	//Sender :
	go func() {
		i := 0
		for {
			time.Sleep(1 * time.Second)
			message := msg.NewCommand("B", "test_command", "test_payload"+fmt.Sprint(i))
			sender.Send(*message)
			i++
			if i >= 10 {
				break
			}
		}
	}()

	//Receiver :
	iterator := msg.NewIterator(s[1].Queue["cmd"])
	go func() {
		for {
			event_rc := destination.Wait("cmd", map[string]string{"name": "test_command"}, 5, iterator)
			if event_rc != nil {
				shoset.Log("B2 message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()
	iterator2 := msg.NewIterator(s[0].Queue["cmd"])
	go func() {
		for {
			event_rc := destination.Wait("cmd", map[string]string{"name": "test_command"}, 5, iterator2)
			if event_rc != nil {
				shoset.Log("B1 message received (Payload) : " + event_rc.GetPayload())
			}
		}
	}()

	// Do someting else while it is sending and receiving messages.

	time.Sleep(3 * time.Second) // Wait for the routing to be established
	utilsForTest.PrintManyShosets(s)
	time.Sleep(20 * time.Second) // Wait for the end of the test

	//select {} //Never return
}
