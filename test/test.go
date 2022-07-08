package main // tests run in the main package

import (
	"context"
	"fmt"
	"os"
	"time"

	// "os"
	// "log"

	"github.com/ditrit/shoset"
)

func loopUntilDone(tick time.Duration, ctx context.Context, callback func()) {
	ticker := time.NewTicker(tick)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("leaving...")
			return
		case t := <-ticker.C:
			fmt.Println("Tick at", t)
			callback()
		}
	}
}

// func shosetClient(logicalName, ShosetType, address string) {
// 	c := shoset.NewShoset(logicalName, ShosetType)
// 	c.Protocol(address, "link")

// 	go func() {
// 		for {
// 			time.Sleep(time.Second * time.Duration(5))
// 		}
// 	}()
// 	/*
// 		go func() {
// 			command := msg.NewCommand("orchestrator", "deploy", "{\"appli\": \"toto\"}")
// 			c.SendCommand(command)
// 			event := msg.NewEvent("bus", "coucou", "ok")
// 			c.SendEvent(event)

// 			events := msg.NewIterator(c.qEvents)
// 			defer events.Close()
// 			rec := c.WaitEvent(events, "bus", "started", 20)
// 			if rec != nil {
// 				fmt.Println(">Received Event: \n%#v\n", *rec)
// 			} else {
// 				fmt.Print("Timeout expired !")
// 			}
// 			events2 := msg.NewIterator(c.qEvents)
// 			defer events.Close()
// 			rec2 := c.WaitEvent(events2, "bus", "starting", 20)
// 			if rec2 != nil {
// 				fmt.Println(">Received Event 2: \n%#v\n", *rec2)
// 			} else {
// 				fmt.Print("Timeout expired  2 !")
// 			}
// 		}()

// 	*/
// 	<-c.Done
// }

// func shosetServer(logicalName, ShosetType, address string) {
// 	s := shoset.NewShoset(logicalName, ShosetType)
// 	err := s.Bind(address)

// 	if err != nil {
// 		fmt.Println("Gandalf server socket can not be created")
// 	}

// 	go func() {
// 		for {
// 			time.Sleep(time.Second * time.Duration(5))
// 		}
// 	}()
// 	/*
// 		go func() {
// 			time.Sleep(time.Second * time.Duration(5))
// 			event := msg.NewEvent("bus", "starting", "ok")
// 			s.SendEvent(event)
// 			time.Sleep(time.Millisecond * time.Duration(200))
// 			event = msg.NewEvent("bus", "started", "ok")
// 			s.SendEvent(event)
// 			command := msg.NewCommand("bus", "register", "{\"topic\": \"toto\"}")
// 			s.SendCommand(command)
// 			reply := msg.NewReply(command, "success", "OK")
// 			s.SendReply(reply)
// 		}()
// 	*/
// 	<-s.Done
// }

// func shosetTest() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	c1 := shoset.NewShoset("c", "c")
// 	c1.Bind("127.0.0.1:8301")

// 	c2 := shoset.NewShoset("c", "c")
// 	c2.Bind("127.0.0.1:8302")

// 	c3 := shoset.NewShoset("c", "c")
// 	c3.Bind("127.0.0.1:8303")

// 	d1 := shoset.NewShoset("d", "a")
// 	d1.Bind("127.0.0.1:8401")

// 	d2 := shoset.NewShoset("d", "a")
// 	d2.Bind("127.0.0.1:8402")

// 	b1 := shoset.NewShoset("b", "c")
// 	b1.Bind("127.0.0.1:8201")
// 	b1.Protocol("127.0.0.1:8302", "link")
// 	b1.Protocol("127.0.0.1:8301", "link")
// 	b1.Protocol("127.0.0.1:8303", "link")
// 	b1.Protocol("127.0.0.1:8401", "link")
// 	b1.Protocol("127.0.0.1:8402", "link")

// 	a1 := shoset.NewShoset("a", "c")
// 	a1.Bind("127.0.0.1:8101")
// 	a1.Protocol("127.0.0.1:8201", "link")

// 	b2 := shoset.NewShoset("b", "c")
// 	b2.Bind("127.0.0.1:8202")
// 	b2.Protocol("127.0.0.1:8301", "link")

// 	b3 := shoset.NewShoset("b", "c")
// 	b3.Bind("127.0.0.1:8203")
// 	b3.Protocol("127.0.0.1:8303", "link")

// 	a2 := shoset.NewShoset("a", "c")
// 	a2.Bind("127.0.0.1:8102")
// 	a2.Protocol("127.0.0.1:8202", "link")

// 	time.Sleep(time.Second * time.Duration(1))
// 	fmt.Println("a1 : ", a1.String())
// 	fmt.Println("a2 : ", a2.String())
// 	fmt.Println("b1 : ", b1.String())
// 	fmt.Println("b2 : ", b2.String())
// 	fmt.Println("b3 : ", b3.String())
// 	fmt.Println("c1 : ", c1.String())
// 	fmt.Println("c2 : ", c2.String())
// 	fmt.Println("c3 : ", c3.String())
// 	fmt.Println("d1 : ", d1.String())
// 	fmt.Println("d2 : ", d2.String())
// 	<-done
// }

// func shosetTestEtoile() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	cl1 := shoset.NewShoset("cl", "cl")
// 	cl1.Bind("127.0.0.1:8001")

// 	cl2 := shoset.NewShoset("cl", "cl")
// 	cl2.Bind("127.0.0.1:8002")
// 	cl2.Protocol("127.0.0.1:8001", "join")
// 	cl3 := shoset.NewShoset("cl", "cl")
// 	cl3.Bind("127.0.0.1:8003")
// 	cl3.Protocol("127.0.0.1:8002", "join")

// 	cl4 := shoset.NewShoset("cl", "cl")
// 	cl4.Bind("127.0.0.1:8004")
// 	cl4.Protocol("127.0.0.1:8001", "join")

// 	cl5 := shoset.NewShoset("cl", "cl")
// 	cl5.Bind("127.0.0.1:8005")
// 	cl5.Protocol("127.0.0.1:8001", "join")

// 	aga1 := shoset.NewShoset("aga", "a")
// 	aga1.Bind("127.0.0.1:8111")
// 	aga1.Protocol("127.0.0.1:8001", "link")
// 	aga2 := shoset.NewShoset("aga", "a")
// 	aga2.Bind("127.0.0.1:8112")
// 	aga2.Protocol("127.0.0.1:8005", "link")

// 	agb1 := shoset.NewShoset("agb", "a")
// 	agb1.Bind("127.0.0.1:8121")
// 	agb1.Protocol("127.0.0.1:8002", "link")
// 	agb2 := shoset.NewShoset("agb", "a")
// 	agb2.Bind("127.0.0.1:8122")
// 	agb2.Protocol("127.0.0.1:8003", "link")

// 	time.Sleep(time.Second * time.Duration(2))

// 	Ca1 := shoset.NewShoset("Ca", "c")
// 	Ca1.Bind("127.0.0.1:8211")
// 	Ca1.Protocol("127.0.0.1:8111", "link")
// 	Ca2 := shoset.NewShoset("Ca", "c")
// 	Ca2.Bind("127.0.0.1:8212")
// 	Ca2.Protocol("127.0.0.1:8111", "link")
// 	Ca3 := shoset.NewShoset("Ca", "c")
// 	Ca3.Bind("127.0.0.1:8213")
// 	Ca3.Protocol("127.0.0.1:8111", "link")

// 	Cb1 := shoset.NewShoset("Cb", "c")
// 	Cb1.Bind("127.0.0.1:8221")
// 	Cb1.Protocol("127.0.0.1:8112", "link")
// 	Cb2 := shoset.NewShoset("Cb", "c")
// 	Cb2.Bind("127.0.0.1:8222")
// 	Cb2.Protocol("127.0.0.1:8112", "link")

// 	Cc1 := shoset.NewShoset("Cc", "c")
// 	Cc1.Bind("127.0.0.1:8231")
// 	Cc1.Protocol("127.0.0.1:8111", "link")
// 	Cc2 := shoset.NewShoset("Cc", "c")
// 	Cc2.Bind("127.0.0.1:8232")
// 	Cc2.Protocol("127.0.0.1:8111", "link")

// 	Cd1 := shoset.NewShoset("Cd", "c")
// 	Cd1.Bind("127.0.0.1:8241")
// 	Cd1.Protocol("127.0.0.1:8111", "link")
// 	Cd2 := shoset.NewShoset("Cd", "c")
// 	Cd2.Bind("127.0.0.1:8242")
// 	Cd2.Protocol("127.0.0.1:8112", "link")

// 	Ce1 := shoset.NewShoset("Ce", "c")
// 	Ce1.Bind("127.0.0.1:8251")
// 	Ce1.Protocol("127.0.0.1:8122", "link")
// 	Ce2 := shoset.NewShoset("Ce", "c")
// 	Ce2.Bind("127.0.0.1:8252")
// 	Ce2.Protocol("127.0.0.1:8122", "link")

// 	Cf1 := shoset.NewShoset("Cf", "c")
// 	Cf1.Bind("127.0.0.1:8261")
// 	Cf1.Protocol("127.0.0.1:8121", "link")
// 	Cf2 := shoset.NewShoset("Cg", "c")
// 	Cf2.Bind("127.0.0.1:8262")
// 	Cf2.Protocol("127.0.0.1:8121", "link")

// 	Cg1 := shoset.NewShoset("Cg", "c")
// 	Cg1.Bind("127.0.0.1:8271")
// 	Cg1.Protocol("127.0.0.1:8121", "link")
// 	Cg2 := shoset.NewShoset("Cg", "c")
// 	Cg2.Bind("127.0.0.1:8272")
// 	Cg2.Protocol("127.0.0.1:8122", "link")

// 	Ch1 := shoset.NewShoset("Ch", "c")
// 	Ch1.Bind("127.0.0.1:8281")
// 	Ch1.Protocol("127.0.0.1:8111", "link")

// 	time.Sleep(time.Second * time.Duration(2))
// 	fmt.Println("cl1 : ", cl2.String())
// 	fmt.Println("cl2 : ", cl2.String())
// 	fmt.Println("cl3 : ", cl3.String())
// 	fmt.Println("cl4 : ", cl4.String())
// 	fmt.Println("cl5 : ", cl5.String())

// 	fmt.Println("aga1 : ", aga1.String())
// 	fmt.Println("aga2 : ", aga2.String())

// 	fmt.Println("agb1 : ", agb1.String())
// 	fmt.Println("agb2 : ", agb2.String())

// 	fmt.Println("Ca1 : ", Ca1.String())
// 	fmt.Println("Ca2 : ", Ca2.String())
// 	fmt.Println("Ca3 : ", Ca3.String())

// 	fmt.Println("Cb1 : ", Cb1.String())
// 	fmt.Println("Cb2 : ", Cb2.String())

// 	fmt.Println("Cc1 : ", Cc1.String())
// 	fmt.Println("Cc2 : ", Cc2.String())

// 	fmt.Println("Cd1 : ", Cd1.String())
// 	fmt.Println("Cd2 : ", Cd2.String())

// 	fmt.Println("Ce1 : ", Ce1.String())
// 	fmt.Println("Ce2 : ", Ce2.String())

// 	fmt.Println("Cf1 : ", Cf1.String())
// 	fmt.Println("Cf2 : ", Cf2.String())

// 	fmt.Println("Cg1 : ", Cg1.String())
// 	fmt.Println("Cg2 : ", Cg2.String())

// 	fmt.Println("Ch1 : ", Ch1.String())

// 	<-done
// }

// func testQueue() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
//
//      /*      // First let's make 2 sockets talk each other
// 		C1 := shoset.NewShoset("C1", "c")
// 		C1.Bind("127.0.0.1:8261")
// 		C1.Protocol("127.0.0.1:8262","link")

// 		C2 := shoset.NewShoset("C2", "cl")
// 		C2.Bind("127.0.0.1:8262")
// 		C2.Protocol("127.0.0.1:8261","link")

// 		// Let's check for sockets connections
// 		time.Sleep(time.Second * time.Duration(1))

// 		fmt.Println("C1 : ", C1.String())
// 		fmt.Println("C2 : ", C2.String())

// 		// Make C1 send a message to C2
// 		socket := C1.GetConnByAddr(C2.GetBindAddr())
// 		m := msg.NewCommand("test", "test", "content")
// 		m.Timeout = 10000
// 		fmt.Println("Message Pushed: %+v\n", *m)
// 		socket.SendMessage(m)

// 		// Let's dump C2 queue for cmd msg
// 		time.Sleep(time.Second * time.Duration(1))
// 		cell := C2.FQueue("cmd").First()
// 		fmt.Println("Cell in queue: %+v\n", *cell)
// 	*/<-done
// }

func simpleCluster() {
	done := make(chan bool)
	cl1 := shoset.NewShoset("cl", "cl")
	cl1.InitPKI("127.0.0.1:8001") //we take the port 8001 for our first socket
	for {
		time.Sleep(time.Second * time.Duration(5))
		fmt.Println("\ncl : ", cl1)
	}
	<-done
}

func simpleAgregator() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", aga1)
	})
}

func simpleConnector() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Ca1 := shoset.NewShoset("Ca", "c") // agregateur
	Ca1.Protocol("127.0.0.1:8211", "127.0.0.1:8111", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", Ca1)
	})
}

func simplesimpleConnector() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Ca1 := shoset.NewShoset("Ca", "c") // agregateur
	Ca1.Bind("127.0.0.1:8211")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", Ca1)
	})
}

func testJoin1() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl1 := shoset.NewShoset("cl", "cl")
	cl1.Bind("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")                      // always "cl" "cl" for gandalf
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join") // we join it to our first socket

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8001", "join")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
	})
}

func testJoin2() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl2 := shoset.NewShoset("cl", "cl")                      // always "cl" "cl" for gandalf
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join") // we join it to our first socket

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8001", "join")
	// cl3.Protocol("127.0.0.1:8002", "join")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
	})
}

func testJoin3(ctx context.Context, done context.CancelFunc) {
	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.InitPKI("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("c", "c")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "link")

	cl5 := shoset.NewShoset("cl", "cl")
	cl5.Protocol("127.0.0.1:8005", "127.0.0.1:8003", "join")

	cl6 := shoset.NewShoset("c", "c")
	cl6.Protocol("127.0.0.1:8006", "127.0.0.1:8002", "link")

	time.Sleep(time.Second * time.Duration(2))
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8003", "bye")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Printf("%s: %v", cl1.GetLogicalName(), cl1)
		fmt.Printf("%s: %v", cl2.GetLogicalName(), cl2)
		fmt.Printf("%s: %v", cl3.GetLogicalName(), cl3)
		fmt.Printf("%s: %v", cl4.GetLogicalName(), cl4)
		fmt.Printf("%s: %v", cl5.GetLogicalName(), cl5)
		fmt.Printf("%s: %v", cl6.GetLogicalName(), cl6)
		done()
	})
}

func testJoin4() {
	done := make(chan bool)

	cl2 := shoset.NewShoset("cl", "cl")                      // always "cl" "cl" for gandalf
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join") // we join it to our first socket

	cl3 := shoset.NewShoset("a", "a")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "link")

	cl4 := shoset.NewShoset("a", "a")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "link") // we join it to our first socket

	cl5 := shoset.NewShoset("b", "b")
	cl5.Protocol("127.0.0.1:8005", "127.0.0.1:8004", "link")

	for {
		time.Sleep(time.Second * time.Duration(5))
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\ncl : ", cl5)
	}

	<-done
}

// func testJoin() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	cl2 := shoset.NewShoset("cl", "cl") // always "cl" "cl" for gandalf
// 	fmt.Println("\ncl : ", cl2)
// 	cl2.Bind("127.0.0.1:8002") //we take the port 8002 for our first socket
// 	cl2.Protocol("127.0.0.1:8001", "join") // we join it to our first socket

// 	cl3 := shoset.NewShoset("cl", "cl")
// 	cl3.Bind("127.0.0.1:8003")
// 	cl3.Protocol("127.0.0.1:8001", "join")
// 	cl3.Protocol("127.0.0.1:8002", "join")

// 	cl4 := shoset.NewShoset("cl", "cl")
// 	cl4.Bind("127.0.0.1:8004")
// 	cl4.Protocol("127.0.0.1:8002", "join") // we join it to our first socket

// 	for {
// 		time.Sleep(time.Second * time.Duration(2))
// 		fmt.Println("\ncl : ", cl2)
// 		fmt.Println("\ncl : ", cl3)
// 		fmt.Println("\ncl : ", cl4)
// 	}

// 	<-done
// }

// func testLink() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	cl1 := shoset.NewShoset("cl", "cl") // cluster
// 	cl1.Bind("127.0.0.1:8001")

// 	cl2 := shoset.NewShoset("cl", "cl")
// 	cl2.Bind("127.0.0.1:8002")
// 	cl2.Protocol("127.0.0.1:8001","join")

// 	cl3 := shoset.NewShoset("cl", "cl")
// 	cl3.Bind("127.0.0.1:8003")
// 	cl3.Protocol("127.0.0.1:8002","join")

// 	aga1 := shoset.NewShoset("aga", "a") // agregateur
// 	aga1.Bind("127.0.0.1:8111")
// 	aga1.Protocol("127.0.0.1:8001","link")

// 	aga2 := shoset.NewShoset("aga", "a") // agregateur
// 	aga2.Bind("127.0.0.1:8112")
// 	aga2.Protocol("127.0.0.1:8002","link")

// 	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
// 	Ca1.Bind("127.0.0.1:8211")
// 	Ca1.Protocol("127.0.0.1:8111","link")

// 	time.Sleep(time.Second * time.Duration(3))
// 	aga3 := shoset.NewShoset("aga", "a") // agregateur
// 	aga3.Bind("127.0.0.1:8113")
// 	aga3.Protocol("127.0.0.1:8002","link")

// 	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
// 	Ca2.Bind("127.0.0.1:8212")
// 	Ca2.Protocol("127.0.0.1:8113","link")

// 	for {
// 		fmt.Println("\ncl : ", cl1)
// 		fmt.Println("\ncl : ", cl2)
// 		// fmt.Println("\n", cl2.ConnsByLname)
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

func testLink1() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8001", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	})
}

func testLink2() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8001", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	})
}

func testLink3() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8002", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	})
}

func testLink4() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8002", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
	})
}

func testLink5() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8002", "link")

	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
	Ca1.Protocol("127.0.0.1:8211", "127.0.0.1:8111", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Protocol("127.0.0.1:8212", "127.0.0.1:8112", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca1)
		fmt.Println("\nca : ", Ca2)
	})
}

func testLink6() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8002", "link")

	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
	Ca1.Protocol("127.0.0.1:8211", "127.0.0.1:8111", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Protocol("127.0.0.1:8212", "127.0.0.1:8112", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga1)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca1)
		fmt.Println("\nca : ", Ca2)
	})
}

func testLink7() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.Bind("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8002", "link")

	Ca1 := shoset.NewShoset("Ca", "c") //connecteur
	Ca1.Protocol("127.0.0.1:8211", "127.0.0.1:8111", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Protocol("127.0.0.1:8212", "127.0.0.1:8112", "link")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("\ncl : ", cl1)
		fmt.Println("\ncl : ", cl2)
		fmt.Println("\ncl : ", cl3)
		fmt.Println("\ncl : ", cl4)
		fmt.Println("\nag : ", aga2)
		fmt.Println("\nca : ", Ca1)
		fmt.Println("\nca : ", Ca2)
	})
}

func testLink8() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.InitPKI("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	aga1 := shoset.NewShoset("aga", "a") // agregateur
	aga1.Protocol("127.0.0.1:8111", "127.0.0.1:8001", "link")

	aga2 := shoset.NewShoset("aga", "a") // agregateur
	aga2.Protocol("127.0.0.1:8112", "127.0.0.1:8002", "link")

	Ca2 := shoset.NewShoset("Ca", "c") //connecteur
	Ca2.Protocol("127.0.0.1:8212", "127.0.0.1:8112", "link")

	// time.Sleep(time.Second * time.Duration(5))

	// fmt.Println("\ncl : ", cl1)
	// fmt.Println("\ncl : ", cl2)
	// fmt.Println("\ncl : ", cl3)
	// fmt.Println("\ncl : ", cl4)
	// fmt.Println("\nag : ", aga1)
	// fmt.Println("\nag : ", aga2)
	// fmt.Println("\nca : ", Ca2)

	// time.Sleep(time.Second * time.Duration(5))

	// cl1.Protocol("127.0.0.1:8001", "bye")

	// time.Sleep(time.Second * time.Duration(1))

	// fmt.Println("\ncl : ", cl1)
	// fmt.Println("\ncl : ", cl2)
	// fmt.Println("\ncl : ", cl3)
	// fmt.Println("\ncl : ", cl4)
	// fmt.Println("\nag : ", aga1)
	// fmt.Println("\nag : ", aga2)
	// fmt.Println("\nca : ", Ca2)

	loopUntilDone(2*time.Second, ctx, func() {
		// fmt.Println("\ncl : ", cl1)
		// fmt.Println("\ncl : ", cl2)
		// fmt.Println("\ncl : ", cl3)
		// fmt.Println("\ncl : ", cl4)
		// fmt.Println("\nag : ", aga1)
		// fmt.Println("\nag : ", aga2)
		// fmt.Println("\nca : ", Ca2)
		// fmt.Println("ConnsByTypeArray('cl')", aga1.GetConnsByTypeArray("c"))
	})
}

func testPki(ctx context.Context, done context.CancelFunc) {
	tt := []struct {
		lname, stype, src, dst, ptype string
	}{
		{lname: "cl", stype: "cl", src: "127.0.0.1:8001", dst: "", ptype: "pki"},
		{lname: "cl", stype: "cl", src: "127.0.0.1:8002", dst: "127.0.0.1:8001", ptype: "join"},
		{lname: "cl", stype: "cl", src: "127.0.0.1:8003", dst: "127.0.0.1:8002", ptype: "join"},
		{lname: "cl", stype: "cl", src: "127.0.0.1:8004", dst: "127.0.0.1:8001", ptype: "join"},
		{lname: "aga", stype: "a", src: "127.0.0.1:8111", dst: "127.0.0.1:8001", ptype: "link"},
		{lname: "aga", stype: "a", src: "127.0.0.1:8112", dst: "127.0.0.1:8002", ptype: "link"},
		{lname: "Ca", stype: "c", src: "127.0.0.1:8211", dst: "127.0.0.1:8111", ptype: "link"},
		{lname: "Ca", stype: "c", src: "127.0.0.1:8212", dst: "127.0.0.1:8112", ptype: "link"},
		{lname: "w", stype: "w", src: "127.0.0.1:8311", dst: "127.0.0.1:8211", ptype: "link"},
		{lname: "x", stype: "x", src: "127.0.0.1:8312", dst: "127.0.0.1:8212", ptype: "link"},
		{lname: "y", stype: "y", src: "127.0.0.1:8412", dst: "127.0.0.1:8312", ptype: "link"},
		{lname: "z", stype: "z", src: "127.0.0.1:8512", dst: "127.0.0.1:8412", ptype: "link"},
	}

	s := make([]*shoset.Shoset, len(tt))
	for i, t := range tt {
		s[i] = shoset.NewShoset(t.lname, t.stype)
		if t.ptype == "pki" {
			s[i].InitPKI(t.src)
		} else {
			s[i].Protocol(t.src, t.dst, t.ptype)
		}
	}

	// time.Sleep(time.Second * time.Duration(2))
	// s[2].Protocol("127.0.0.1:8003", "127.0.0.1:8002", "bye")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Println("in_callback")
		for _, conn := range s {
			fmt.Printf("%s: %v", conn.GetLogicalName(), conn)
		}
		done()
	})
}

func testPkiServer(ctx context.Context, done context.CancelFunc) {
	cl1 := shoset.NewShoset("cl", "cl") // cluster
	cl1.InitPKI("127.0.0.1:8001")

	loopUntilDone(2*time.Second, ctx, func() {
		// fmt.Println("\ncl : ", cl1)
		done()
		return
	})
}

func testPkiClient(ctx context.Context, done context.CancelFunc) {
	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("x", "x")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "link")

	cl4 := shoset.NewShoset("y", "y")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8003", "link")

	loopUntilDone(2*time.Second, ctx, func() {
		// fmt.Println("\ncl : ", cl2)
		done()
	})
}

func testPresentationENIB(ctx context.Context, done context.CancelFunc) {
	cl1 := shoset.NewShoset("cl", "cl")
	cl1.InitPKI("127.0.0.1:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("127.0.0.1:8002", "127.0.0.1:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("127.0.0.1:8003", "127.0.0.1:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("127.0.0.1:8004", "127.0.0.1:8001", "join")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Printf("%s: %v", cl1.GetLogicalName(), cl1)
		fmt.Printf("%s: %v", cl2.GetLogicalName(), cl2)
		fmt.Printf("%s: %v", cl3.GetLogicalName(), cl3)
		fmt.Printf("%s: %v", cl4.GetLogicalName(), cl4)
		done()
	})
}

func main() {
	shoset.InitPrettyLogger(false)
	shoset.SetLogLevel("info")

	ctx, done := context.WithTimeout(context.Background(), 1*time.Minute)

	//terminal
	arg := os.Args[1]
	if arg == "1" {
		shoset.Log("testPkiServer")
		testPkiServer(ctx, done)
		// testJoin1()
		// testJoin2()
		// testJoin3()
		// testJoin4()
		// testLink1()
		// testLink2()
		// testLink3()
		// testLink4()
		// testLink5()
		// testLink6()
		// testLink7()
		// testLink8()
		// testPki()
	} else if arg == "2" {
		shoset.Log("testPkiClient")
		// testPkiClient(ctx, done)
		simpleCluster()
		// simpleAgregator()
		// simpleConnector()
	} else if arg == "3" {
		shoset.Log("simplesimpleConnector")
		// simplesimpleConnector()
	} else {
		shoset.Log("testPki")
		testPki(ctx, done)
		// testPresentationENIB(ctx, done)
		// testJoin4()
	}
}

// linkOk
