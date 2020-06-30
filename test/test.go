package main

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
)

const (
	ca = `
-----BEGIN CERTIFICATE-----
MIIFbzCCA1egAwIBAgIUCXRC33qPMen5yUng+oUTpo+Vh1AwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCRlIxEzARBgNVBAgMClNvbWUgU3RhdGUxEjAQBgNVBAcM
CVRoZSBDbG91ZDEPMA0GA1UECgwGT3JuZXNzMB4XDTIwMDYxNTExMjgwNFoXDTMw
MDYxMzExMjgwNFowRzELMAkGA1UEBhMCRlIxEzARBgNVBAgMClNvbWUgU3RhdGUx
EjAQBgNVBAcMCVRoZSBDbG91ZDEPMA0GA1UECgwGT3JuZXNzMIICIjANBgkqhkiG
9w0BAQEFAAOCAg8AMIICCgKCAgEA75EFN3UgbP0UvZsJS9RH6FFCHXmWSfmL/dU5
LmEByBXUxLakkZ8AmP4tz1zPv0EHIWJJy6mRDqpEJGKQv9LDxAyI5rWo63XniSwI
3/Wr+2jK+UgoG1PBbZ6lMxhclI1B4E03mG+5aRFdMdN1u/Th/2q5Lzo4cMLl/NU1
DYkhXvEg+nD5pAEqN+1wVHQbCZBMX07POGCDHN5rMG7MI0ESyz4VkMJfj9mwDBg7
BB3Jmd1ZkWsh0t5WR+95LEmDwMwLMxHoXuaXsfIr4vRD52WwSxrs1jai9MqQs2gm
/VtM0UM/xK0Nzm4Gl3HdtyXmj4/H7iHaDcGth52E4VG/C1iEejf9LM2BPob9pukM
2E19eq3bYrFcW2xv5y0FcEfzVhJNIqS/ZTICb39yH8mvaWN6VyZSEu5VBSJy5Z5d
0hSfTgKLVwPMkoRcWeau+Q7Iiwiik8j9ceCEb7bL+mwiwaqeIxNAPVVrvUymbE84
ME7WbeVWdZKAGoKcJvH/niKC6XeH1CpL2DcrLlOFq7OM+caO+89dNRFN0nXbEZaR
LbipUlOvwqJ0RXnN/nlUND85qF5MtnKH/tcr0wNjitguztZtpaVxOj+9VUrlyTUv
8DVkvmMo2HiAHUMF8Jrh+VgHfhM8ofSY2lDfVPiaqLN8zCwuZpYgzy/xk19JxI7r
dhPfRDUCAwEAAaNTMFEwHQYDVR0OBBYEFJLLqDsD4y0LsfLksJY5U4NrbZ4iMB8G
A1UdIwQYMBaAFJLLqDsD4y0LsfLksJY5U4NrbZ4iMA8GA1UdEwEB/wQFMAMBAf8w
DQYJKoZIhvcNAQELBQADggIBAGeP1TuhhZEUNn3uMJzHAJhKPNDPsfNmFo7lVDPp
mgX8hwlZTJXQNU27d0/ISA1EnZ+44fBb3rKygXLkOLJ1JOo9tYHlpMk5IOGF70fd
a140xPNwW40pDSkYr0dlGrvRBXi9vvJVuk7FQfsQOrZPNS5GxK55ygXOPU+3TpM7
ZQB1jHX4lcxiJ/zPlx54Ew30yUqX6n6OEYNDvKGaWBEn51PFuntIrfkOV9SDvEos
LKXGSA/81JxKw7arKqzyhqVp4iUXRyRxHDasHjHulkCpB4ddQ8jgw85eEZP6SpA7
NbwixWOJNsXjMtKy29dDhEAi7cIeX3tytGMArUB5k2xMPuBvWcujxLxOiEsa51kU
65qIUHlLr9f41knTHxJGbkKqr/V9VISDRiaic6EhDsJTtpHQveoIJplR++2NSgSj
3wuR+fyB7wqPW9Y3gaDw3sySc5fKDvxpifJoakFpDVNLARdo0k0U+BqGOR35vuRm
LmOUPTmX9cVS5sDJOwm4SS9gm+VnRePaY5Yb5/AuAmGgPdBN3fq0ZGeO7HI7UpR0
K9UR+IzCuScX0/koJlrrni31j2hnX4eTL0tTy3LqaqAw/YlXkgqR392LnDNrp/w5
vMfQY/7bVXgBV3mNN5+tPmIpC/ZRoelhzatfS39uXcGu7tU7H6ngP4Ju0DpJQmIe
qD2n
-----END CERTIFICATE-----`

	cert = `
-----BEGIN CERTIFICATE-----
MIIFoTCCA4mgAwIBAgIUR4/E6KLGkH3rWAvke/hG/2L6lHQwDQYJKoZIhvcNAQEL
BQAwRzELMAkGA1UEBhMCRlIxEzARBgNVBAgMClNvbWUgU3RhdGUxEjAQBgNVBAcM
CVRoZSBDbG91ZDEPMA0GA1UECgwGT3JuZXNzMB4XDTIwMDYxNTEyNTI1OFoXDTIx
MDYyMDEyNTI1OFowYzELMAkGA1UEBhMCRlIxEzARBgNVBAgMClNvbWUgU3RhdGUx
EjAQBgNVBAcMCVRoZSBDbG91ZDEPMA0GA1UECgwGT3JuZXNzMRowGAYDVQQDDBFn
YW5kYWxmLmRpdHJpdC5pbzCCAiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIB
APKp6HToiW3OBLOs37sxPN6unGW3uDA1LcwD6JoBhhIdMDGIg0aqYf//HH2Im6mx
ZcmRFAcx2lr3RjguVRhSbQGrua9aWf2vhwIRX993/IoCRRLyJw1+0ckaIbdumwKS
FwFm9Npp2DR0e7NR7ZWSDZ/BWihkWmZH+ccv2jkr10pAFkSnMvHDSuNeXfU3jc9H
Em+zro/8sKgImLUtd3fLNhS0g+RwInPE6tU8PT3aw6aZwuX9NIRzCsLFsn+Sldso
cSuMCMbXRJcWsrkQBVTF7iJ7P1QoRM9m7q3APXh2mMnEA8p5n+/+fK0P91vdNbLu
g+XyIncDuw7nmuEKF8B4BXTF9WeqxTbdIDXdsQzMomng3YNxWTaX4WnXQ1q3k8A1
SebYXbll9nGgz1MOUWCwrFY1AGp1PSROxg/41JbyNBK4a/ykbizBQi7MTKhFHhcR
PcbtFy34NbF+WbNa33cWXd//AdHsLA7pYuQRlkXBU5iFXuUbdBYipFsccZZAd9w+
kmHOS9JjCTImT5cSRaypCs5B1sE9blvFQS85VdjX8MxN3imxbs2mcJDHaq9zZNdH
ON2bXYMM5inzDnsNN3BP7StdvnQCkxri76DNqZaprPzO5lT8QsNsZbbHsSAg8V6J
xb6EAjXbJWM9IYivGpmSEIysBuwNcvWyy76JGjskOOoXAgMBAAGjaTBnMAkGA1Ud
EwQCMAAwHQYDVR0OBBYEFLbC4NlYdolvIrjSNzg2jldvmyeRMAsGA1UdDwQEAwIF
oDAdBgNVHSUEFjAUBggrBgEFBQcDAgYIKwYBBQUHAwEwDwYDVR0RBAgwBocEfwAA
ATANBgkqhkiG9w0BAQsFAAOCAgEAKGQUI3YT0n2Q+MfKi587jvvCjSegQ49W6FiH
lPmgAI9/T5IvRuROA9tUWSivmkwzsxVpR/8KmSFR0oG8Zs0yU00tR7URQxyZKSI/
U8PVkZ1/VcGHVDrfJjo2rMnIgX/FHhrunVz3kApqpXUfWEbpUtLDqa8BZdEGS1hk
jseaA6+wzNUbgIstY1vd4Eeizx0Zcqzt/28A7w6feYvgCt40Oa0QCi2HwxGCKxEf
5TTEN41wX3M6MRveWaMQDnavOVezf5m9fBzkfWftDMdHvpuVxMnyPzH/csm+U36M
OmV5dOwHdlCyr2Oyfn94/8Ipm527gBbnwDiX3sk5J0IqJGktyhYudghWtO89mOyf
SjqMkOFQQG6xMWwfKmLfyJJoyr2R3fhc1oGldTlhSE8U3xmBZ6r3zbNDYRibA9MQ
5CxxCSX9wwyBcGUuw0OJ1nTHRDoBQ7F1fTYb6x7e9xTOrfa5wYgx1WPH2nxd6ZCm
vA1GTF0pQNcJVoQrStUyBEEeQa/XeqQ6WxAlTFlqb3hMekJIIXPF/VlRICpm9SMe
8NJ64kDlbPjm1OgkmDLny55Sq6PLgm2XOROS3biwa57Dxd6O4UkDRVjR17j8qBtQ
RLZvLLcGa/4A7zgQzG8V1J1EuG0DXpxDafKP03T3A6P/EKkLAFMRFjT3vvxNQdh9
/OQc1LQ=
-----END CERTIFICATE-----`

	key = `
-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
)

var certs = map[string][]byte{
	"ca":    []byte(ca),
	"cakey": nil,
	"cert":  []byte(cert),
	"key":   []byte(key),
}

func shosetClient(logicalName, ShosetType, address string) {
	c := shoset.NewShoset(logicalName, ShosetType, certs)
	c.Link(address)

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
	s := shoset.NewShoset(logicalName, ShosetType, certs)
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

	c1 := shoset.NewShoset("c", "c", certs)
	c1.Bind("localhost:8301")

	c2 := shoset.NewShoset("c", "c", certs)
	c2.Bind("localhost:8302")

	c3 := shoset.NewShoset("c", "c", certs)
	c3.Bind("localhost:8303")

	d1 := shoset.NewShoset("d", "a", certs)
	d1.Bind("localhost:8401")

	d2 := shoset.NewShoset("d", "a", certs)
	d2.Bind("localhost:8402")

	b1 := shoset.NewShoset("b", "c", certs)
	b1.Bind("localhost:8201")
	b1.Link("localhost:8302")
	b1.Link("localhost:8301")
	b1.Link("localhost:8303")
	b1.Link("localhost:8401")
	b1.Link("localhost:8402")

	a1 := shoset.NewShoset("a", "c", certs)
	a1.Bind("localhost:8101")
	a1.Link("localhost:8201")

	b2 := shoset.NewShoset("b", "c", certs)
	b2.Bind("localhost:8202")
	b2.Link("localhost:8301")

	b3 := shoset.NewShoset("b", "c", certs)
	b3.Bind("localhost:8203")
	b3.Link("localhost:8303")

	a2 := shoset.NewShoset("a", "c", certs)
	a2.Bind("localhost:8102")
	a2.Link("localhost:8202")

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

	cl1 := shoset.NewShoset("cl", "cl", certs)
	cl1.Bind("localhost:8001")

	cl2 := shoset.NewShoset("cl", "cl", certs)
	cl2.Bind("localhost:8002")
	cl2.Join("localhost:8001")
	cl3 := shoset.NewShoset("cl", "cl", certs)
	cl3.Bind("localhost:8003")
	cl3.Join("localhost:8002")

	cl4 := shoset.NewShoset("cl", "cl", certs)
	cl4.Bind("localhost:8004")
	cl4.Join("localhost:8001")

	cl5 := shoset.NewShoset("cl", "cl", certs)
	cl5.Bind("localhost:8005")
	cl5.Join("localhost:8001")

	aga1 := shoset.NewShoset("aga", "a", certs)
	aga1.Bind("localhost:8111")
	aga1.Link("localhost:8001")
	aga2 := shoset.NewShoset("aga", "a", certs)
	aga2.Bind("localhost:8112")
	aga2.Link("localhost:8005")

	agb1 := shoset.NewShoset("agb", "a", certs)
	agb1.Bind("localhost:8121")
	agb1.Link("localhost:8002")
	agb2 := shoset.NewShoset("agb", "a", certs)
	agb2.Bind("localhost:8122")
	agb2.Link("localhost:8003")

	time.Sleep(time.Second * time.Duration(2))

	Ca1 := shoset.NewShoset("Ca", "c", certs)
	Ca1.Bind("localhost:8211")
	Ca1.Link("localhost:8111")
	Ca2 := shoset.NewShoset("Ca", "c", certs)
	Ca2.Bind("localhost:8212")
	Ca2.Link("localhost:8111")
	Ca3 := shoset.NewShoset("Ca", "c", certs)
	Ca3.Bind("localhost:8213")
	Ca3.Link("localhost:8111")

	Cb1 := shoset.NewShoset("Cb", "c", certs)
	Cb1.Bind("localhost:8221")
	Cb1.Link("localhost:8112")
	Cb2 := shoset.NewShoset("Cb", "c", certs)
	Cb2.Bind("localhost:8222")
	Cb2.Link("localhost:8112")

	Cc1 := shoset.NewShoset("Cc", "c", certs)
	Cc1.Bind("localhost:8231")
	Cc1.Link("localhost:8111")
	Cc2 := shoset.NewShoset("Cc", "c", certs)
	Cc2.Bind("localhost:8232")
	Cc2.Link("localhost:8111")

	Cd1 := shoset.NewShoset("Cd", "c", certs)
	Cd1.Bind("localhost:8241")
	Cd1.Link("localhost:8111")
	Cd2 := shoset.NewShoset("Cd", "c", certs)
	Cd2.Bind("localhost:8242")
	Cd2.Link("localhost:8112")

	Ce1 := shoset.NewShoset("Ce", "c", certs)
	Ce1.Bind("localhost:8251")
	Ce1.Link("localhost:8122")
	Ce2 := shoset.NewShoset("Ce", "c", certs)
	Ce2.Bind("localhost:8252")
	Ce2.Link("localhost:8122")

	Cf1 := shoset.NewShoset("Cf", "c", certs)
	Cf1.Bind("localhost:8261")
	Cf1.Link("localhost:8121")
	Cf2 := shoset.NewShoset("Cg", "c", certs)
	Cf2.Bind("localhost:8262")
	Cf2.Link("localhost:8121")

	Cg1 := shoset.NewShoset("Cg", "c", certs)
	Cg1.Bind("localhost:8271")
	Cg1.Link("localhost:8121")
	Cg2 := shoset.NewShoset("Cg", "c", certs)
	Cg2.Bind("localhost:8272")
	Cg2.Link("localhost:8122")

	Ch1 := shoset.NewShoset("Ch", "c", certs)
	Ch1.Bind("localhost:8281")
	Ch1.Link("localhost:8111")

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
		C1.Link("localhost:8262")

		C2 := shoset.NewShoset("C2", "cl")
		C2.Bind("localhost:8262")
		C2.Link("localhost:8261")

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

func main() {
	// fmt.Println("Running shosetTest")
	// shosetTest()

	fmt.Println("Running shosetTestEtoile")
	shosetTestEtoile()
}
