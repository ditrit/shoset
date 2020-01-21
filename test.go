package main

import (
	"fmt"
	"time"

	"./net"
)

func shosetClient(logicalName string, address string) {
	c := net.NewShoset(logicalName)
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

func shosetServer(logicalName string, address string) {
	s := net.NewShoset(logicalName)
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

func shosetTest() {
	done := make(chan bool)

	c1 := net.NewShoset("c")
	c1.Bind("localhost:8301")

	c2 := net.NewShoset("c")
	c2.Bind("localhost:8302")

	c3 := net.NewShoset("c")
	c3.Bind("localhost:8303")

	d1 := net.NewShoset("d")
	d1.Bind("localhost:8401")

	d2 := net.NewShoset("d")
	d2.Bind("localhost:8402")

	b1 := net.NewShoset("b")
	b1.Bind("localhost:8201")
	b1.Connect("localhost:8302")
	b1.Connect("localhost:8301")
	b1.Connect("localhost:8303")
	b1.Connect("localhost:8401")
	b1.Connect("localhost:8402")

	a1 := net.NewShoset("a")
	a1.Bind("localhost:8101")
	a1.Connect("localhost:8201")

	b2 := net.NewShoset("b")
	b2.Bind("localhost:8202")
	b2.Connect("localhost:8301")

	b3 := net.NewShoset("b")
	b3.Bind("localhost:8203")
	b3.Connect("localhost:8303")

	a2 := net.NewShoset("a")
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

func shosetTestEtoile() {
	done := make(chan bool)

	cl1 := net.NewShoset("cl")
	cl1.Bind("localhost:8001")

	cl2 := net.NewShoset("cl")
	cl2.Bind("localhost:8002")
	cl2.Join("localhost:8001")
	cl3 := net.NewShoset("cl")
	cl3.Bind("localhost:8003")
	cl3.Join("localhost:8002")

	cl4 := net.NewShoset("cl")
	cl4.Bind("localhost:8004")
	cl4.Join("localhost:8001")

	cl5 := net.NewShoset("cl")
	cl5.Bind("localhost:8005")
	cl5.Join("localhost:8001")

	aga1 := net.NewShoset("aga")
	aga1.Bind("localhost:8111")
	aga1.Connect("localhost:8001")
	aga2 := net.NewShoset("aga")
	aga2.Bind("localhost:8112")
	aga2.Connect("localhost:8005")

	agb1 := net.NewShoset("agb")
	agb1.Bind("localhost:8121")
	agb1.Connect("localhost:8002")
	agb2 := net.NewShoset("agb")
	agb2.Bind("localhost:8122")
	agb2.Connect("localhost:8003")

	time.Sleep(time.Second * time.Duration(2))

	Ca1 := net.NewShoset("Ca")
	Ca1.Bind("localhost:8211")
	Ca1.Connect("localhost:8111")
	Ca2 := net.NewShoset("Ca")
	Ca2.Bind("localhost:8212")
	Ca2.Connect("localhost:8111")
	Ca3 := net.NewShoset("Ca")
	Ca3.Bind("localhost:8213")
	Ca3.Connect("localhost:8111")

	Cb1 := net.NewShoset("Cb")
	Cb1.Bind("localhost:8221")
	Cb1.Connect("localhost:8112")
	Cb2 := net.NewShoset("Cb")
	Cb2.Bind("localhost:8222")
	Cb2.Connect("localhost:8112")

	Cc1 := net.NewShoset("Cc")
	Cc1.Bind("localhost:8231")
	Cc1.Connect("localhost:8111")
	Cc2 := net.NewShoset("Cc")
	Cc2.Bind("localhost:8232")
	Cc2.Connect("localhost:8111")

	Cd1 := net.NewShoset("Cd")
	Cd1.Bind("localhost:8241")
	Cd1.Connect("localhost:8111")
	Cd2 := net.NewShoset("Cd")
	Cd2.Bind("localhost:8242")
	Cd2.Connect("localhost:8112")

	Ce1 := net.NewShoset("Ce")
	Ce1.Bind("localhost:8251")
	Ce1.Connect("localhost:8122")
	Ce2 := net.NewShoset("Ce")
	Ce2.Bind("localhost:8252")
	Ce2.Connect("localhost:8122")

	Cf1 := net.NewShoset("Cf")
	Cf1.Bind("localhost:8261")
	Cf1.Connect("localhost:8121")
	Cf2 := net.NewShoset("Cg")
	Cf2.Bind("localhost:8262")
	Cf2.Connect("localhost:8121")

	Cg1 := net.NewShoset("Cg")
	Cg1.Bind("localhost:8271")
	Cg1.Connect("localhost:8121")
	Cg2 := net.NewShoset("Cg")
	Cg2.Bind("localhost:8272")
	Cg2.Connect("localhost:8122")

	Ch1 := net.NewShoset("Ch")
	Ch1.Bind("localhost:8281")
	Ch1.Connect("localhost:8111")

	time.Sleep(time.Second * time.Duration(2))
	fmt.Printf("cl1 : %s", cl2.String())
	fmt.Printf("cl2 : %s", cl2.String())
	fmt.Printf("cl3 : %s", cl3.String())
	fmt.Printf("cl4 : %s", cl4.String())
	fmt.Printf("cl5 : %s", cl5.String())

	fmt.Printf("aga1 : %s", aga1.String())
	fmt.Printf("aga2 : %s", aga2.String())

	fmt.Printf("agb1 : %s", agb1.String())
	fmt.Printf("agb2 : %s", agb2.String())

	fmt.Printf("Ca1 : %s", Ca1.String())
	fmt.Printf("Ca2 : %s", Ca2.String())
	fmt.Printf("Ca3 : %s", Ca3.String())

	fmt.Printf("Cb1 : %s", Cb1.String())
	fmt.Printf("Cb2 : %s", Cb2.String())

	fmt.Printf("Cc1 : %s", Cc1.String())
	fmt.Printf("Cc2 : %s", Cc2.String())

	fmt.Printf("Cd1 : %s", Cd1.String())
	fmt.Printf("Cd2 : %s", Cd2.String())

	fmt.Printf("Ce1 : %s", Ce1.String())
	fmt.Printf("Ce2 : %s", Ce2.String())

	fmt.Printf("Cf1 : %s", Cf1.String())
	fmt.Printf("Cf2 : %s", Cf2.String())

	fmt.Printf("Cg1 : %s", Cg1.String())
	fmt.Printf("Cg2 : %s", Cg2.String())

	fmt.Printf("Ch1 : %s", Ch1.String())

	<-done
}
