package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"os"

	"./msg"
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

	flag.BoolVar(&isServer, "s", false, "Server mode (shorthand)")
	flag.BoolVar(&isServer, "server", false, "Server mode")
	flag.BoolVar(&isClient, "c", false, "Client mode (shorthand)")
	flag.BoolVar(&isClient, "client", false, "Client mode")
	flag.Parse()

	args := flag.Args()

	if (isServer == isClient) || (len(args) != 2) {
		flag.Usage()
		os.Exit(1)
	}

	gob.Register(new(msg.Event))
	gob.Register(new(msg.Command))
	gob.Register(new(msg.Reply))
	gob.Register(new(msg.Config))

	name := args[0]
	address := args[1]

	if isServer == true {
		server(name, address)
	}
	if isClient == true {
		client(name, address)
	}
}
