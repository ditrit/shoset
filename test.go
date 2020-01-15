package main

import (
	"fmt"
	"time"

	"./net"
)

func chaussetteClient(logicalName string, address string) {
	c := net.NewChaussette(logicalName)
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
	<-c.Done
}

func chaussetteServer(logicalName string, address string) {
	s := net.NewChaussette(logicalName)
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
	<-s.Done
}

func chaussetteTest() {
	done := make(chan bool)

	c1 := net.NewChaussette("c")
	c1.Bind("localhost:8301")

	c2 := net.NewChaussette("c")
	c2.Bind("localhost:8302")

	c3 := net.NewChaussette("c")
	c3.Bind("localhost:8303")

	d1 := net.NewChaussette("d")
	d1.Bind("localhost:8401")

	d2 := net.NewChaussette("d")
	d2.Bind("localhost:8402")

	b1 := net.NewChaussette("b")
	b1.Bind("localhost:8201")
	b1.Connect("localhost:8302")
	b1.Connect("localhost:8301")
	b1.Connect("localhost:8303")
	b1.Connect("localhost:8401")
	b1.Connect("localhost:8402")

	a1 := net.NewChaussette("a")
	a1.Bind("localhost:8101")
	a1.Connect("localhost:8201")

	b2 := net.NewChaussette("b")
	b2.Bind("localhost:8202")
	b2.Connect("localhost:8301")

	b3 := net.NewChaussette("b")
	b3.Bind("localhost:8203")
	b3.Connect("localhost:8303")

	a2 := net.NewChaussette("a")
	a2.Bind("localhost:8102")
	a2.Connect("localhost:8202")

	time.Sleep(time.Second * time.Duration(1))
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
