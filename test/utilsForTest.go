package main // tests run in the main package

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
	//"github.com/ditrit/shoset/msg"
)

type ShosetCreation struct {
	lname    string
	stype    string
	src      string
	dst      []string
	ptype    string
	launched bool
}

func createManyShosets(tt []*ShosetCreation, s []*shoset.Shoset, wait bool) []*shoset.Shoset {
	for i, t := range tt {
		if !t.launched {
			s = append(s, shoset.NewShoset(t.lname, t.stype))
			//s[i] = shoset.NewShoset(t.lname, t.stype)
			if t.ptype == "pki" {
				s[i].InitPKI(t.src)
			} else {
				for _, a := range t.dst {
					s[i].Protocol(t.src, a, t.ptype)
				}
			}
			t.launched = true
		}
	}
	if wait {
		time.Sleep(1 * time.Second)
	}

	return s
}

func printManyShosets(s []*shoset.Shoset) {
	for i, t := range s {
		fmt.Println("\nShoset ", i, ": ", t)
	}
}

func WaitForManyShosets(s []*shoset.Shoset) {
	for _, t := range s {
		t.WaitForProtocols()
	}
}

func routeManyShosets(s []*shoset.Shoset, wait bool) {
	for _, t := range s {
		routing := msg.NewRoutingEvent(t.GetLogicalName(), "")
		t.Send(routing)
	}
	if wait {
		time.Sleep(1 * time.Second)
	}
}
