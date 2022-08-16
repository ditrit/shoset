package main // tests run in the main package

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	// "os"
	// "log"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
	example "github.com/ditrit/shoset/test/example"
	oldTest "github.com/ditrit/shoset/test/old_test"
	utilsForTest "github.com/ditrit/shoset/test/utils_for_test"
)

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

	utilsForTest.LoopUntilDone(1*time.Second, ctx, func() {
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

	utilsForTest.LoopUntilDone(2*time.Second, ctx, func() {
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

	utilsForTest.LoopUntilDone(2*time.Second, ctx, func() {
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

	utilsForTest.LoopUntilDone(1*time.Second, ctx, func() {
		fmt.Printf("%s: %v", cl1.GetLogicalName(), cl1)
		fmt.Printf("%s: %v", cl2.GetLogicalName(), cl2)
		fmt.Printf("%s: %v", cl3.GetLogicalName(), cl3)
		fmt.Printf("%s: %v", cl4.GetLogicalName(), cl4)
		done()
	})
}

// #### Routing test

func testRouteTable(ctx context.Context, done context.CancelFunc) {

	tt := utilsForTest.StraightLine // Choose the network topology for the test
	s := []*shoset.Shoset{}
	s = utilsForTest.CreateManyShosets(tt, s, false)
	utilsForTest.WaitForManyShosets(s)

	time.Sleep(10 * time.Second)

	utilsForTest.PrintManyShosets(s)

	tt = append(tt, &(utilsForTest.ShosetCreation{Lname: "F", ShosetType: "cl", LocalAddress: "localhost:8006", RemoteAddresses: []string{"localhost:8001", "localhost:8005"}, ProtocolType: "link", Launched: false}))

	s = utilsForTest.CreateManyShosets(tt, s, false)
	utilsForTest.WaitForManyShosets(s)

	time.Sleep(10 * time.Second)

	utilsForTest.PrintManyShosets(s)
}

func testForwardMessage(ctx context.Context, done context.CancelFunc) {
	tt := utilsForTest.Circle
	s := []*shoset.Shoset{}
	s = utilsForTest.CreateManyShosets(tt, s, false)
	utilsForTest.WaitForManyShosets(s)

	time.Sleep(1 * time.Second)

	utilsForTest.PrintManyShosets(s)

	var wg sync.WaitGroup

	destination := s[len(s)-1]

	// Receive Message
	wg.Add(1)
	go func() {
		defer wg.Done()
		event_rc := destination.Wait("simpleMessage", map[string]string{}, 30, nil)
		fmt.Println("(Main) Message received : ", event_rc)
	}()

	// Send Message
	message := msg.NewSimpleMessage(destination.GetLogicalName(), "test_payload")
	fmt.Println("Message sent : ", message)
	s[0].Send(message)

	wg.Wait()
}

func testForwardMessageMultiProcess2(args []string) {
	// args[0] is not the nama of the execuatable, it is the first argument after test number
	// ./bin/shoset_build 5 D 0 0 rien 0 (args[0] is D)
	fmt.Println("args : ", args)

	// To generate profiles and traces about only one shoset
	// if args[0] != "D" {

	// 	//Disable tracer and profiles in the main.

	// 	// Tracer
	// 	f, _ := os.Create("./profiler/trace_" + args[0] + ".out")
	// 	defer f.Close()
	// 	trace.Start(f)
	// 	defer trace.Stop()

	// 	// CPU profiler
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

	var cl *shoset.Shoset
	if args[4] == "1" {
		fmt.Println("#### Relaunch")
		time.Sleep(2 * time.Second)
		cl = utilsForTest.CreateShosetOnlyBindFromTopology(args[0], utilsForTest.StraightLine)
	} else {
		fmt.Println("#### Launch")
		cl = utilsForTest.CreateShosetFromTopology(args[0], utilsForTest.StraightLine)
	}

	fmt.Println("Waiting for protocols to complete.")
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
				time.Sleep(1 * time.Second)
			}
		}()
	}

	fmt.Println("#### Shoset ", cl.GetLogicalName(), "is ready.")

	for {
		time.Sleep(10 * time.Second)

		fmt.Println("Shoset ", cl.GetLogicalName(), " : ", cl)
	}

	//select {}
}

// Send an event every second forever :
func testEndConnection(ctx context.Context, done context.CancelFunc) {
	tt := utilsForTest.Line3 // Choose the network topology for the test

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
			fmt.Println("Message sent : ", message)
			sender.Send(message)
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

	time.Sleep(5 * time.Second)

	fmt.Println("####", s[2].GetLogicalName(), " is ending connection to B")
	s[2].EndProtocol("B", "127.0.0.1:8002")

	time.Sleep(5 * time.Second)

	utilsForTest.PrintManyShosets(s)

	time.Sleep(10 * time.Second)
}

func main() {
	// Clear the content of the profiler folder
	// os.RemoveAll("./profiler/")
	// os.MkdirAll("./profiler/", 0777)

	// tracer
	// f, _ := os.Create("./profiler/trace.out")
	// defer f.Close()
	// trace.Start(f)
	// defer trace.Stop()

	// CPU profiler
	// var cpuprofile = flag.String("cpuprofile", "./profiler/cpu.prof", "write cpu profile to `file`")
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
	shoset.SetLogLevel(shoset.TRACE)

	ctx, done := context.WithTimeout(context.Background(), 1*time.Minute)

	//terminal
	// Choose the test to run, only decomment one for each case.
	arg := os.Args[1]
	if arg == "1" {
		shoset.Log("testPkiServer")
		// testPkiServer(ctx, done)
		oldTest.TestJoin1()
		// oldTest.TestJoin2()
		// oldTest.TestJoin3()
		// oldTest.TestJoin4()
		// oldTest.TestLink1()
		// oldTest.TestLink2()
		// oldTest.TestLink3()
		// oldTest.TestLink4()
		// oldTest.TestLink5()
		// oldTest.TestLink6()
		// oldTest.TestLink7()
		// oldTest.TestLink8()
		// testPki()
	} else if arg == "2" {
		shoset.Log("testPkiClient")
		// testPkiClient(ctx, done)
		// oldTest.SimpleCluster()
		// oldTest.SimpleAgregator()
		// oldTest.SimpleConnector()
	} else if arg == "3" {
		shoset.Log("simplesimpleConnector")
		//oldTest.SimplesimpleConnector()
	} else if arg == "4" {
		// testPki(ctx, done)
		// testPresentationENIB(ctx, done)
		// oldTest.TestJoin3(ctx, done)

		// testRouteTable(ctx, done)
		testForwardMessage(ctx, done)
		// testEndConnection(ctx, done)
	} else if arg == "5" {
		testForwardMessageMultiProcess2((os.Args)[2:])
	} else if arg == "6" {
		example.SimpleExample()
		// example.TestEventContinuousSend()
		// example.TestSimpleForwarding()
		// example.TestForwardingTopology()
	}

	// Memory profiler
	// var memprofile = flag.String("memprofile", "./profiler/mem.prof", "write memory profile to `file`")

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
