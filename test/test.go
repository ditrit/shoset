package main // tests run in the main package

import (
	"fmt"
	"os"

	"time"

	"github.com/ditrit/shoset"
)

func shosetClient(logicalName, ShosetType, address string) {
	c := shoset.NewShoset(logicalName, ShosetType)
	c.Protocol(address, "link")

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

func shosetServer(logicalName, ShosetType, address string) {
	s := shoset.NewShoset(logicalName, ShosetType)
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

	c1 := shoset.NewShoset("c", "c")
	c1.Bind("localhost:8301")

	c2 := shoset.NewShoset("c", "c")
	c2.Bind("localhost:8302")

	c3 := shoset.NewShoset("c", "c")
	c3.Bind("localhost:8303")

	d1 := shoset.NewShoset("d", "a")
	d1.Bind("localhost:8401")

	d2 := shoset.NewShoset("d", "a")
	d2.Bind("localhost:8402")

	b1 := shoset.NewShoset("b", "c")
	b1.Bind("localhost:8201")
	b1.Protocol("localhost:8302", "link")
	b1.Protocol("localhost:8301", "link")
	b1.Protocol("localhost:8303", "link")
	b1.Protocol("localhost:8401", "link")
	b1.Protocol("localhost:8402", "link")

	a1 := shoset.NewShoset("a", "c")
	a1.Bind("localhost:8101")
	a1.Protocol("localhost:8201", "link")

	b2 := shoset.NewShoset("b", "c")
	b2.Bind("localhost:8202")
	b2.Protocol("localhost:8301", "link")

	b3 := shoset.NewShoset("b", "c")
	b3.Bind("localhost:8203")
	b3.Protocol("localhost:8303", "link")

	a2 := shoset.NewShoset("a", "c")
	a2.Bind("localhost:8102")
	a2.Protocol("localhost:8202", "link")

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

	cl1 := shoset.NewShoset("cl", "cl")
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")
	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Bind("localhost:8004")
	cl4.Protocol("localhost:8001", "join")

	cl5 := shoset.NewShoset("cl", "cl")
	cl5.Bind("localhost:8005")
	cl5.Protocol("localhost:8001", "join")

	aga1 := shoset.NewShoset("aga", "a")
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")
	aga2 := shoset.NewShoset("aga", "a")
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8005", "link")

	agb1 := shoset.NewShoset("agb", "a")
	agb1.Bind("localhost:8121")
	agb1.Protocol("localhost:8002", "link")
	agb2 := shoset.NewShoset("agb", "a")
	agb2.Bind("localhost:8122")
	agb2.Protocol("localhost:8003", "link")

	time.Sleep(time.Second * time.Duration(2))

	Ca1 := shoset.NewShoset("Ca", "c")
	Ca1.Bind("localhost:8211")
	Ca1.Protocol("localhost:8111", "link")
	Ca2 := shoset.NewShoset("Ca", "c")
	Ca2.Bind("localhost:8212")
	Ca2.Protocol("localhost:8111", "link")
	Ca3 := shoset.NewShoset("Ca", "c")
	Ca3.Bind("localhost:8213")
	Ca3.Protocol("localhost:8111", "link")

	Cb1 := shoset.NewShoset("Cb", "c")
	Cb1.Bind("localhost:8221")
	Cb1.Protocol("localhost:8112", "link")
	Cb2 := shoset.NewShoset("Cb", "c")
	Cb2.Bind("localhost:8222")
	Cb2.Protocol("localhost:8112", "link")

	Cc1 := shoset.NewShoset("Cc", "c")
	Cc1.Bind("localhost:8231")
	Cc1.Protocol("localhost:8111", "link")
	Cc2 := shoset.NewShoset("Cc", "c")
	Cc2.Bind("localhost:8232")
	Cc2.Protocol("localhost:8111", "link")

	Cd1 := shoset.NewShoset("Cd", "c")
	Cd1.Bind("localhost:8241")
	Cd1.Protocol("localhost:8111", "link")
	Cd2 := shoset.NewShoset("Cd", "c")
	Cd2.Bind("localhost:8242")
	Cd2.Protocol("localhost:8112", "link")

	Ce1 := shoset.NewShoset("Ce", "c")
	Ce1.Bind("localhost:8251")
	Ce1.Protocol("localhost:8122", "link")
	Ce2 := shoset.NewShoset("Ce", "c")
	Ce2.Bind("localhost:8252")
	Ce2.Protocol("localhost:8122", "link")

	Cf1 := shoset.NewShoset("Cf", "c")
	Cf1.Bind("localhost:8261")
	Cf1.Protocol("localhost:8121", "link")
	Cf2 := shoset.NewShoset("Cg", "c")
	Cf2.Bind("localhost:8262")
	Cf2.Protocol("localhost:8121", "link")

	Cg1 := shoset.NewShoset("Cg", "c")
	Cg1.Bind("localhost:8271")
	Cg1.Protocol("localhost:8121", "link")
	Cg2 := shoset.NewShoset("Cg", "c")
	Cg2.Bind("localhost:8272")
	Cg2.Protocol("localhost:8122", "link")

	Ch1 := shoset.NewShoset("Ch", "c")
	Ch1.Bind("localhost:8281")
	Ch1.Protocol("localhost:8111", "link")

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

func testQueue() {
	done := make(chan bool)
	/*	// First let's make 2 sockets talk each other
		C1 := shoset.NewShoset("C1", "c")
		C1.Bind("localhost:8261")
		C1.Protocol("localhost:8262","link")

		C2 := shoset.NewShoset("C2", "cl")
		C2.Bind("localhost:8262")
		C2.Protocol("localhost:8261","link")

		// Let's check for sockets connections
		time.Sleep(time.Second * time.Duration(1))

		fmt.Printf("C1 : %s", C1.String())
		fmt.Printf("C2 : %s", C2.String())

		// Make C1 send a message to C2
		socket := C1.GetConnByAddr(C2.GetBindAddr())
		m := msg.NewCommand("test", "test", "content")
		m.Timeout = 10000
		fmt.Printf("Message Pushed: %+v\n", *m)
		socket.SendMessage(m)

		// Let's dump C2 queue for cmd msg
		time.Sleep(time.Second * time.Duration(1))
		cell := C2.FQueue("cmd").First()
		fmt.Printf("Cell in queue: %+v\n", *cell)
	*/<-done
}

func simpleCluster() {
	done := make(chan bool)
	cl1 := shoset.NewShoset("cl", "cl")
	cl1.Bind("localhost:8001") //we take the port 8001 for our first socket
	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
	}
	<-done
}

func simpleAgregator() {
	done := make(chan bool)
	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", aga1)
	}
	<-done
}

func simpleConnector() {
	done := make(chan bool)
	Ca1 := shoset.NewShoset("Ca", "c") // agregateur
	Ca1.Bind("localhost:8211")
	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", Ca1)
	}
	<-done
}

func testJoin1() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl")
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")    // always "cl" "cl" for gandalf
	cl2.Bind("localhost:8002")             //we take the port 8002 for our first socket
	cl2.Protocol("localhost:8001", "join") // we join it to our first socket

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8001", "join")
	// cl3.Protocol("localhost:8002", "join")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
	}

	<-done
}

func testJoin2() {
	done := make(chan bool)

	cl2 := shoset.NewShoset("cl", "cl")    // always "cl" "cl" for gandalf
	cl2.Bind("localhost:8002")             //we take the port 8002 for our first socket
	cl2.Protocol("localhost:8001", "join") // we join it to our first socket

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8001", "join")
	// cl3.Protocol("localhost:8002", "join")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
	}

	<-done
}

func testJoin3() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	cl5 := shoset.NewShoset("cl", "cl")
	cl5.Bind("localhost:8005")
	cl5.Protocol("localhost:8001", "join")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl5)
	}

	<-done
}

func testJoin4() {
	done := make(chan bool)

	cl2 := shoset.NewShoset("cl", "cl")    // always "cl" "cl" for gandalf
	cl2.Bind("localhost:8002")             //we take the port 8002 for our first socket
	cl2.Protocol("localhost:8001", "join") // we join it to our first socket

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	// cl3.Protocol("localhost:8001", "join")
	cl3.Protocol("localhost:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Bind("localhost:8004")
	cl4.Protocol("localhost:8001", "join") // we join it to our first socket

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
	}

	<-done
}

// func testJoin() {
// 	done := make(chan bool)

// 	cl2 := shoset.NewShoset("cl", "cl") // always "cl" "cl" for gandalf
// 	fmt.Println("\ncl : ", cl2)
// 	cl2.Bind("localhost:8002") //we take the port 8002 for our first socket
// 	cl2.Protocol("localhost:8001", "join") // we join it to our first socket

// 	cl3 := shoset.NewShoset("cl", "cl")
// 	cl3.Bind("localhost:8003")
// 	cl3.Protocol("localhost:8001", "join")
// 	cl3.Protocol("localhost:8002", "join")

// 	cl4 := shoset.NewShoset("cl", "cl")
// 	cl4.Bind("localhost:8004")
// 	cl4.Protocol("localhost:8002", "join") // we join it to our first socket

// 	for {
// 		time.Sleep(time.Second * time.Duration(2))
// 		fmt.Println("\ncl : ", cl2)
// 		fmt.Println("\ncl : ", cl3)
// 		fmt.Println("\ncl : ", cl4)
// 	}

// 	<-done
// }

// func test_link() {
// 	done := make(chan bool)

// 	cl1 := shoset.NewShoset("cl", "cl") // cluster
// 	cl1.Bind("localhost:8001")

// 	cl2 := shoset.NewShoset("cl", "cl")
// 	cl2.Bind("localhost:8002")
// 	cl2.Protocol("localhost:8001","join")

// 	cl3 := shoset.NewShoset("cl", "cl")
// 	cl3.Bind("localhost:8003")
// 	cl3.Protocol("localhost:8002","join")

// 	aga1 := shoset.NewShoset("aga", "a") // agregateur
// 	aga1.Bind("localhost:8111")
// 	aga1.Protocol("localhost:8001","link")

// 	aga2 := shoset.NewShoset("aga", "a") // agregateur
// 	aga2.Bind("localhost:8112")
// 	aga2.Protocol("localhost:8002","link")

// 	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
// 	Ca1.Bind("localhost:8211")
// 	Ca1.Protocol("localhost:8111","link")

// 	time.Sleep(time.Second * time.Duration(3))
// 	aga3 := shoset.NewShoset("aga", "a") // agregateur
// 	aga3.Bind("localhost:8113")
// 	aga3.Protocol("localhost:8002","link")

// 	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
// 	Ca2.Bind("localhost:8212")
// 	Ca2.Protocol("localhost:8113","link")

// 	for {
// 		fmt.Println("\ncl : ", cl1)
// 		fmt.Println("\ncl : ", cl2)
// 		// fmt.Println("\n", cl2.ConnsByName)
// 		fmt.Println("\ncl : ", cl3)
// 		// fmt.Println("\nag : ", aga1)
// 		fmt.Println("\nag : ", aga2)
// 		fmt.Println("\nag : ", aga3)
// 		fmt.Println("\nca : ", Ca1)
// 		fmt.Println("\nca : ", Ca2)
// 		time.Sleep(time.Second * time.Duration(3))
// 	}

// 	<-done
// }

func test_link1() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8001", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	}

	<-done
}

func test_link2() {
	done := make(chan bool)

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8001", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	}

	<-done
}

func test_link3() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8002", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	}

	<-done
}

func test_link4() {
	done := make(chan bool)

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8002", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	}

	<-done
}

func test_link5() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	cl5 := shoset.NewShoset("cl", "cl")
	cl5.Bind("localhost:8005")
	cl5.Protocol("localhost:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8002", "link")

	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
	Ca1.Bind("localhost:8211")
	Ca1.Protocol("localhost:8111", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Bind("localhost:8212")
	Ca2.Protocol("localhost:8112", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1) 
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl5)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca1)
		fmt.Println("\nca : ", Ca2)
	}

	<-done
}

func test_link6() {
	done := make(chan bool)

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8002", "link")

	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
	Ca1.Bind("localhost:8211")
	Ca1.Protocol("localhost:8111", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Bind("localhost:8212")
	Ca2.Protocol("localhost:8112", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca1)
		fmt.Println("\nca : ", Ca2)
	}

	<-done
}

func test_link7() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8002", "link")

	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
	Ca1.Bind("localhost:8211")
	Ca1.Protocol("localhost:8111", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Bind("localhost:8212")
	Ca2.Protocol("localhost:8112", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca1)
		fmt.Println("\nca : ", Ca2)
	}

	<-done
}

func test_link8() {
	done := make(chan bool)

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Bind("localhost:8002")
	cl2.Protocol("localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Bind("localhost:8003")
	cl3.Protocol("localhost:8002", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Bind("localhost:8111")
	aga1.Protocol("localhost:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Bind("localhost:8112")
	aga2.Protocol("localhost:8002", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Bind("localhost:8212")
	Ca2.Protocol("localhost:8112", "link")

	for {
		time.Sleep(time.Second * time.Duration(1))
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca2)
	}

	<-done
}

func main() {
	//terminal
	arg := os.Args[1]
	fmt.Println(arg)
	if arg == "1" {
		// testJoin1()
		// testJoin2()
		// testJoin3()
		// testJoin4()
		// test_link1()
		// test_link2()
		// test_link3()
		// test_link4()
		test_link5()
		// test_link6()
		// test_link7()
		// test_link8()
	} else if arg == "2" {
		simpleCluster()
		// simpleAgregator()
		// simpleConnector()
	} else {
		fmt.Println("You must specify one parameter")
	}
}

// linkOk