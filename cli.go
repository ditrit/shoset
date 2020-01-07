package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("	tcp [options] name ipaddr:port\n options:\n")
		flag.PrintDefaults()
		fmt.Printf("  server and client modes are exclusive\n")
		fmt.Printf(" arguments:\n")
		fmt.Printf("  name	       logical name used for routing\n")
		fmt.Printf("  ipaddr:port  address to bind / connect the socket \n")
	}

	var isServer bool
	var isClient bool
	var isTest bool

	flag.BoolVar(&isServer, "s", false, "Server mode (shorthand)")
	flag.BoolVar(&isServer, "server", false, "Server mode")
	flag.BoolVar(&isClient, "c", false, "Client mode (shorthand)")
	flag.BoolVar(&isClient, "client", false, "Client mode")
	flag.BoolVar(&isTest, "t", false, "Test mode (shorthand)")
	flag.BoolVar(&isTest, "test", false, "Test mode")
	flag.Parse()

	args := flag.Args()

	if isTest == true {
		chaussetteTest()
		return
	}

	if (isServer == isClient) || (len(args) != 2) {
		flag.Usage()
		os.Exit(1)
	}

	name := args[0]
	address := args[1]

	if isServer == true {
		chaussetteServer(name, address)
	}
	if isClient == true {
		chaussetteClient(name, address)
	}

}
