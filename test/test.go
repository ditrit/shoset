package main // tests run in the main package

import (
	"context"
	"fmt"
	"os"
	"runtime/trace"
	"sync"
	"time"

	// "os"
	// "log"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
	utilsForTest "github.com/ditrit/shoset/test/utils_for_test"
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

func testPki(ctx context.Context, done context.CancelFunc) {
	tt := []struct {
		lname, stype, src, dst, ptype string
	}{
		{lname: "cl", stype: "cl", src: "localhost:8001", dst: "", ptype: "pki"},
		{lname: "cl", stype: "cl", src: "localhost:8002", dst: "localhost:8001", ptype: "join"},
		{lname: "cl", stype: "cl", src: "localhost:8003", dst: "localhost:8002", ptype: "join"},
		{lname: "cl", stype: "cl", src: "localhost:8004", dst: "localhost:8001", ptype: "join"},
		{lname: "aga", stype: "a", src: "localhost:8111", dst: "localhost:8001", ptype: "link"},
		{lname: "aga", stype: "a", src: "localhost:8112", dst: "localhost:8002", ptype: "link"},
		{lname: "Ca", stype: "c", src: "localhost:8211", dst: "localhost:8111", ptype: "link"},
		{lname: "Ca", stype: "c", src: "localhost:8212", dst: "localhost:8112", ptype: "link"},
		{lname: "w", stype: "w", src: "localhost:8311", dst: "localhost:8211", ptype: "link"},
		{lname: "x", stype: "x", src: "localhost:8312", dst: "localhost:8212", ptype: "link"},
		{lname: "y", stype: "y", src: "localhost:8412", dst: "localhost:8312", ptype: "link"},
		{lname: "z", stype: "z", src: "localhost:8512", dst: "localhost:8412", ptype: "link"},
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
	// s[2].Protocol("localhost:8003", "localhost:8002", "bye")

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
	cl1.InitPKI("localhost:8001")

	loopUntilDone(2*time.Second, ctx, func() {
		// fmt.Println("\ncl : ", cl1)
		done()
		return
	})
}

func testPkiClient(ctx context.Context, done context.CancelFunc) {
	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("localhost:8002", "localhost:8001", "join")

	cl3 := shoset.NewShoset("x", "x")
	cl3.Protocol("localhost:8003", "localhost:8002", "link")

	cl4 := shoset.NewShoset("y", "y")
	cl4.Protocol("localhost:8004", "localhost:8003", "link")

	loopUntilDone(2*time.Second, ctx, func() {
		// fmt.Println("\ncl : ", cl2)
		done()
	})
}

func testPresentationENIB(ctx context.Context, done context.CancelFunc) {
	cl1 := shoset.NewShoset("cl", "cl")
	cl1.InitPKI("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl")
	cl2.Protocol("localhost:8002", "localhost:8001", "join")

	cl3 := shoset.NewShoset("cl", "cl")
	cl3.Protocol("localhost:8003", "localhost:8002", "join")

	cl4 := shoset.NewShoset("cl", "cl")
	cl4.Protocol("localhost:8004", "localhost:8001", "join")

	loopUntilDone(1*time.Second, ctx, func() {
		fmt.Printf("%s: %v", cl1.GetLogicalName(), cl1)
		fmt.Printf("%s: %v", cl2.GetLogicalName(), cl2)
		fmt.Printf("%s: %v", cl3.GetLogicalName(), cl3)
		fmt.Printf("%s: %v", cl4.GetLogicalName(), cl4)
		done()
	})
}

func testRouteTable(ctx context.Context, done context.CancelFunc) {

	tt := utilsForTest.StraightLine

	s := []*shoset.Shoset{}

	s = utilsForTest.CreateManyShosets(tt, s, true)

	utilsForTest.RouteManyShosets(s, true)

	// routing := msg.NewRoutingEvent("A", "")
	// s[0].Send(routing)

	//time.Sleep(5 * time.Second)

	utilsForTest.PrintManyShosets(s)

	tt = append(tt, &(utilsForTest.ShosetCreation{Lname: "F", Stype: "cl", Src: "localhost:8006", Dst: []string{"localhost:8001", "localhost:8005"}, Ptype: "link", Launched: false}))

	s = utilsForTest.CreateManyShosets(tt, s, true)

	utilsForTest.RouteManyShosets(s, true)

	utilsForTest.PrintManyShosets(s)
}

func testForwardMessage(ctx context.Context, done context.CancelFunc) {

	tt := utilsForTest.Circle // Choose the network topology for the test

	s := []*shoset.Shoset{}

	s = utilsForTest.CreateManyShosets(tt, s, false)

	//time.Sleep(20*time.Second)

	utilsForTest.WaitForManyShosets(s)

	utilsForTest.PrintManyShosets(s)

	var wg sync.WaitGroup

	destination := s[len(s)-1] //.GetLogicalName()

	// Receive Message
	wg.Add(1)
	go func() {
		defer wg.Done()
		//time.Sleep(1 * time.Second)
		event_rc := destination.Wait("simpleMessage", map[string]string{}, 30, nil)
		fmt.Println("(Main) Message received : ", event_rc)
	}()

	// Send Message
	//time.Sleep(1 * time.Second)
	message := msg.NewSimpleMessage(destination.GetLogicalName(), "test_payload")
	message.Timeout = 10000
	fmt.Println("Message sent : ", message)
	s[0].Send(message)

	wg.Wait()

	//PrintManyShosets(s)
}

func testSendEvent() {
	tt := utilsForTest.Simple

	s := []*shoset.Shoset{}

	s = utilsForTest.CreateManyShosets(tt, s, false)

	//fmt.Println("TEST !!")

	s[0].WaitForProtocols(10)
	s[1].WaitForProtocols(10)

	utilsForTest.PrintManyShosets(s)

	//var wg sync.WaitGroup

	// Receive Message

	iterator := msg.NewIterator(s[1].Queue["evt"])

	go func() {
		for { //i := 0; i < 10; i++
			//wg.Add(1)
			//defer wg.Done()

			event_rc := s[1].Wait("evt", map[string]string{"topic": "test_topic", "event": "test_event"}, 10, iterator)
			fmt.Println("(main) Message received : ", event_rc)
		}
	}()

	// Send Message
	go func() {
		for {
			message := msg.NewEventClassic("test_topic", "test_event", "test_payload")
			fmt.Println("Message sent : ", message)
			s[0].Send(message)
			// Timing minimal pour que la gestion de la réception puisse s'éxécuter
			time.Sleep(1 * time.Second)
		}
	}()

	time.Sleep(3 * time.Second)

	s[1].Protocol("localhost:8002", "localhost:8001", "bye")

	time.Sleep(10 * time.Second)

	utilsForTest.PrintManyShosets(s)

	//wg.Wait()
}

func testForwardMessageMultiProcess(args []string) {
	fmt.Println("args : ", args)
	cl := shoset.NewShoset(args[0], "cl") //args[0] : lname
	if args[1] == "1" {                   // args[1] : master
		cl.InitPKI(args[2]) // args[2] : IP
	} else {
		cl.Protocol(args[2], args[3], "link") // args[2] : IP , args[3] : remote IP for connexion
	}

	cl.WaitForProtocols(10)

	// Receive Message
	if args[6] == "1" { //args[5] receiver
		fmt.Println("Receiver : ", cl.GetLogicalName())
		//for {
		event_rc := cl.Wait("simpleMessage", map[string]string{}, 10, msg.NewIterator(cl.Queue["simpleMessage"]))
		fmt.Println("(main) Message received : ", event_rc)
		time.Sleep(10 * time.Millisecond)
		//}
	}

	// Send Message
	if args[4] == "1" { //args[4] sender
		//time.Sleep(1 * time.Second)
		fmt.Println("Sender : ", cl.GetLogicalName())
		message := msg.NewSimpleMessage(args[5], "test_payload "+cl.GetLogicalName())
		fmt.Println("Message sent : ", message)
		cl.Send(message)
	}

	fmt.Println("DONE !!")

	time.Sleep(5 * time.Second)
}

func testForwardMessageMultiProcess2(args []string) {
	fmt.Println("args : ", args)

	// if args[0] != "D" {
	// 	f, _ := os.Create("./profiler/trace_" + args[0] + ".out")
	// 	defer f.Close()
	// 	trace.Start(f)
	// 	defer trace.Stop()

	// 	var cpuprofile = flag.String("cpuprofile", "./profiler/cpu_"+args[0]+".prof", "write cpu profile to `file`")

	// 	flag.Parse()
	// 	if *cpuprofile != "" {
	// 		f, err := os.Create(*cpuprofile)
	// 		if err != nil {
	// 			log.Fatal("could not create CPU profile: ", err)
	// 		}
	// 		defer f.Close() // error handling omitted for example
	// 		if err := pprof.StartCPUProfile(f); err != nil {
	// 			log.Fatal("could not start CPU profile: ", err)
	// 		}
	// 		defer pprof.StopCPUProfile()
	// 	}
	// }

	cl := utilsForTest.CreateShosetFromTopology(args[0], utilsForTest.StraightLine)

	fmt.Println("Waiting for protocols to complete !!")
	cl.WaitForProtocols(10)
	fmt.Println("Shoset : ", cl)

	// Receive Message
	if args[1] == "1" { //args[1] receiver
		fmt.Println("Receiver : ", cl.GetLogicalName())
		iterator := msg.NewIterator(cl.Queue["simpleMessage"])
		go func() {
			for {
				event_rc := cl.Wait("simpleMessage", map[string]string{}, 10, iterator)
				fmt.Println("(main) Message received : ", event_rc)
				//time.Sleep(10 * time.Millisecond)
			}
		}()

	}

	// Send Message
	if args[2] == "1" { //args[2] sender
		go func() {
			for {
				fmt.Println("Sender : ", cl.GetLogicalName())
				message := msg.NewSimpleMessage(args[3], "test_payload "+cl.GetLogicalName()) //args[3] destination
				fmt.Println("Message sent : ", message)
				cl.Send(message)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	fmt.Println("DONE !!")

	// go func() {
	// 	time.Sleep(6 * time.Second)
	// 	panic(nil)
	// }()

	time.Sleep(10 * time.Second)
	fmt.Println("Shoset : ", cl)

	time.Sleep(20 * time.Second)

	select {}

	//panic(nil)
}

func testRelaunch(args []string) {
	fmt.Println("args : ", args)

	f, _ := os.Create("./profiler/trace_" + args[0] + ".out")
	defer f.Close()
	trace.Start(f)
	defer trace.Stop()

	// Create Shoset
	cl := utilsForTest.CreateShosetOnlyBindFromTopology(args[0], utilsForTest.StraightLine)

	fmt.Println("Waiting for protocols to complete !!")
	cl.WaitForProtocols(10)

	fmt.Println("Shoset : ", cl)

	// Receive Message
	if args[1] == "1" { //args[1] receiver
		fmt.Println("Receiver : ", cl.GetLogicalName())
		iterator := msg.NewIterator(cl.Queue["simpleMessage"])
		go func() {
			for {
				event_rc := cl.Wait("simpleMessage", map[string]string{}, 10, iterator)
				fmt.Println("(main) Message received : ", event_rc)
				//time.Sleep(10 * time.Millisecond)
			}
		}()

	}

	// Send Message
	if args[2] == "1" { //args[2] sender
		go func() {
			for {
				fmt.Println("Sender : ", cl.GetLogicalName())
				message := msg.NewSimpleMessage(args[3], "test_payload "+cl.GetLogicalName()) //args[3] destination
				fmt.Println("Message sent : ", message)
				cl.Send(message)
				time.Sleep(5 * time.Second)
			}
		}()
	}

	fmt.Println("DONE !!")

	time.Sleep(15 * time.Second)
	fmt.Println("Shoset : ", cl)

	time.Sleep(20 * time.Second)

	select {}
}

func main() {
	// var cpuprofile = flag.String("cpuprofile", "./profiler/cpu.prof", "write cpu profile to `file`")
	// // var memprofile = flag.String("memprofile", "mem.prof", "write memory profile to `file`")

	// os.RemoveAll("./profiler/")
	// os.MkdirAll("./profiler/", 0777)

	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create CPU profile: ", err)
	// 	}
	// 	defer f.Close() // error handling omitted for example
	// 	if err := pprof.StartCPUProfile(f); err != nil {
	// 		log.Fatal("could not start CPU profile: ", err)
	// 	}
	// 	defer pprof.StopCPUProfile()
	// }

	shoset.InitPrettyLogger(true)
	shoset.SetLogLevel("debug")

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
		testPkiClient(ctx, done)
		// simpleCluster()
		// simpleAgregator()
		// simpleConnector()
	} else if arg == "3" {
		shoset.Log("simplesimpleConnector")
		// simplesimpleConnector()
	} else if arg == "4" {
		// testPki(ctx, done)
		// testPresentationENIB(ctx, done)
		// testJoin3(ctx, done)
		//testRouteTable(ctx, done)
		testForwardMessage(ctx, done)
		//testSendEvent()
	} else if arg == "5" {
		//testForwardMessageMultiProcess((os.Args)[2:])
		testForwardMessageMultiProcess2((os.Args)[2:])

	} else if arg == "6" {
		testRelaunch((os.Args)[2:])
	}
	// if *memprofile != "" {
	// 	f, err := os.Create(*memprofile)
	// 	if err != nil {
	// 		log.Fatal("could not create memory profile: ", err)
	// 	}
	// 	defer f.Close() // error handling omitted for example
	// 	runtime.GC()    // get up-to-date statistics
	// 	if err := pprof.WriteHeapProfile(f); err != nil {
	// 		log.Fatal("could not write memory profile: ", err)
	// 	}
	// }
}

// linkOk
