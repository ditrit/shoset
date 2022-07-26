package utilsForTest

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

func TEST() {
	fmt.Println("TEST")
}

func CreateManyShosets(tt []*ShosetCreation, s []*shoset.Shoset, wait bool) []*shoset.Shoset {
	for i, t := range tt {
		if !t.Launched {
			s = append(s, shoset.NewShoset(t.Lname, t.Stype))
			//s[i] = shoset.NewShoset(t.lname, t.stype)
			if t.Ptype == "pki" {
				s[i].InitPKI(t.Src)
			} else {
				for _, a := range t.Dst {
					s[i].Protocol(t.Src, a, t.Ptype)
				}
			}
			t.Launched = true
		}
	}
	if wait {
		time.Sleep(1 * time.Second)
	}

	return s
}

func PrintManyShosets(s []*shoset.Shoset) {
	for i, t := range s {
		fmt.Println("\nShoset ", i, ": ", t)
	}
}

func WaitForManyShosets(s []*shoset.Shoset) {
	for _, t := range s {
		t.WaitForProtocols()
	}
}

// Not used
func RouteManyShosets(s []*shoset.Shoset, wait bool) {
	for _, t := range s {
		routing := msg.NewRoutingEvent(t.GetLogicalName(), "")
		t.Send(routing)
	}
	if wait {
		time.Sleep(1 * time.Second)
	}
}
