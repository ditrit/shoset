package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		fmt.Printf("	tcp [options] ipaddr:port\n options:\n")
		flag.PrintDefaults()
		fmt.Printf("  server and client modes are exclusive\n")
	}

	var isServer bool
	var isClient bool

	flag.BoolVar(&isServer, "s", false, "Server mode (shorthand)")
	flag.BoolVar(&isServer, "server", false, "Server mode")
	flag.BoolVar(&isClient, "c", false, "Client mode (shorthand)")
	flag.BoolVar(&isClient, "client", false, "Client mode")
	flag.Parse()

	args := flag.Args()

	if (isServer == isClient) || (len(args) != 1) {
		flag.Usage()
		os.Exit(1)
	}

	address := args[0]

	if isServer == true {
		server(address)
	}
	if isClient == true {
		client(address)
	}
}
