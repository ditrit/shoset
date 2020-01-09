package main

import (
	"fmt"
	"time"
)

func chaussetteClient(logicalName string, address string) {
	c := NewChaussette(logicalName)
	c.Connect(address)
	time.Sleep(time.Second * time.Duration(1))
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(5))
		}
	}()
	/*
		go func() {
			command := msg.NewCommand("orchestrator", "deploy", "{\"appli\": \"toto\"}")
			c.SendCommand(command)
			event := msg.NewEvent("bus", "coucou", "ok")
			c.SendEvent(event)

			events := msg.NewIterator(c.qEvents)
			defer events.Close()
			rec := c.WaitEvent(events, "bus", "started", 20)
			if rec != nil {
				fmt.Printf(">Received Event: \n%#v\n", *rec)
			} else {
				fmt.Print("Timeout expired !")
			}
			events2 := msg.NewIterator(c.qEvents)
			defer events.Close()
			rec2 := c.WaitEvent(events2, "bus", "starting", 20)
			if rec2 != nil {
				fmt.Printf(">Received Event 2: \n%#v\n", *rec2)
			} else {
				fmt.Print("Timeout expired  2 !")
			}
		}()

	*/
	<-c.done
}

func chaussetteServer(logicalName string, address string) {
	s := NewChaussette(logicalName)
	err := s.Bind(address)
	if err != nil {
		fmt.Println("Gandalf server socket can not be created")
	}
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(5))
		}
	}()
	/*
		go func() {
			time.Sleep(time.Second * time.Duration(5))
			event := msg.NewEvent("bus", "starting", "ok")
			s.SendEvent(event)
			time.Sleep(time.Millisecond * time.Duration(200))
			event = msg.NewEvent("bus", "started", "ok")
			s.SendEvent(event)
			command := msg.NewCommand("bus", "register", "{\"topic\": \"toto\"}")
			s.SendCommand(command)
			reply := msg.NewReply(command, "success", "OK")
			s.SendReply(reply)
		}()
	*/
	<-s.done
}

func chaussetteTest() {
	done := make(chan bool)

	fmt.Printf("\n--\ncreation c1\n")
	c1 := NewChaussette("c")
	c1.Bind("localhost:8301")

	fmt.Printf("\n--\ncreation c2\n")
	c2 := NewChaussette("c")
	c2.Bind("localhost:8302")

	fmt.Printf("\n--\ncreation c3\n")
	c3 := NewChaussette("c")
	c3.Bind("localhost:8303")

	fmt.Printf("\n--\ncreation d1\n")
	d1 := NewChaussette("d")
	d1.Bind("localhost:8401")

	fmt.Printf("\n--\ncreation d2\n")
	d2 := NewChaussette("d")
	d2.Bind("localhost:8402")

	fmt.Printf("\n--\ncreation b1\n")
	b1 := NewChaussette("b")
	b1.Bind("localhost:8201")
	b1.Connect("localhost:8302")
	b1.Connect("localhost:8301")
	b1.Connect("localhost:8303")
	b1.Connect("localhost:8401")
	b1.Connect("localhost:8402")

	fmt.Printf("\n--\ncreation a1\n")
	a1 := NewChaussette("a")
	a1.Bind("localhost:8101")
	a1.Connect("localhost:8201")

	fmt.Printf("\n--\ncreation b2\n")
	b2 := NewChaussette("b")
	b2.Bind("localhost:8202")
	b2.Connect("localhost:8301")

	fmt.Printf("\n--\ncreation b3\n")
	b3 := NewChaussette("b")
	b3.Bind("localhost:8203")
	b3.Connect("localhost:8303")

	fmt.Printf("\n--\ncreation a1\n")
	a2 := NewChaussette("a")
	a2.Bind("localhost:8102")
	a2.Connect("localhost:8202")

	time.Sleep(time.Second * time.Duration(3))
	fmt.Printf("a1 : %s", a1.String())
	fmt.Printf("a2 : %s", a2.String())
	fmt.Printf("b1 : %s", b1.String())
	fmt.Printf("b2 : %s", b2.String())
	fmt.Printf("b3 : %s", b3.String())

	fmt.Printf("c1 : %s", c1.String())
	fmt.Printf("c2 : %s", c2.String())
	fmt.Printf("c3 : %s", c3.String())
	fmt.Printf("d1 : %s", d1.String())
	fmt.Printf("d2 : %s", d2.String())
	<-done
}