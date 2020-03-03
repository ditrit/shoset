package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	shosetTestEtoile()

	return

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("	tcp [options] name ipaddr:port\n options:\n")
		flag.PrintDefaults()
		fmt.Printf("  server and client modes are exclusive\n")
		fmt.Printf(" arguments:\n")
		fmt.Printf("  name	       logical name used for routing\n")
		fmt.Printf("  ShosetType    node type used for routing\n")
		fmt.Printf("  ipaddr:port  address to bind / connect the socket \n")
	}

	var isServer bool
	var isClient bool
	var isTest bool
	var isTestEtoile bool
	var isQueueTest bool

	flag.BoolVar(&isServer, "s", false, "Server mode (shorthand)")
	flag.BoolVar(&isServer, "server", false, "Server mode")
	flag.BoolVar(&isClient, "c", false, "Client mode (shorthand)")
	flag.BoolVar(&isClient, "client", false, "Client mode")
	flag.BoolVar(&isTest, "t", false, "Test mode (shorthand)")
	flag.BoolVar(&isTest, "test", false, "Test mode")
	flag.BoolVar(&isTestEtoile, "test2", false, "Test etoile mode")
	flag.BoolVar(&isTestEtoile, "t2", false, "Test etoile mode")
	flag.BoolVar(&isQueueTest, "q", false, "queue test mode")
	flag.Parse()

	args := flag.Args()

	if isTest == true {
		shosetTest()
		return
	}

	if isTestEtoile == true {
		shosetTestEtoile()
		return
	}

	if isQueueTest == true {
		testQueue()
		return
	}

	if (isServer == isClient) || (len(args) != 3) {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println(args)

	name := args[0]
	ShosetType := args[1]
	address := args[2]

	if isServer == true {
		shosetServer(name, ShosetType, address)
	}
	if isClient == true {
		shosetClient(name, ShosetType, address)
	}

}
