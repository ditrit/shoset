package utilsForTest

import (
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

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

func CreateShosetFromTopology(Lname string, tt []*ShosetCreation) *shoset.Shoset {
	for _, t := range tt {
		if !t.Launched && t.Lname == Lname {
			s := shoset.NewShoset(t.Lname, t.Stype)
			if t.Ptype == "pki" {
				s.InitPKI(t.Src)
			} else {
				for _, a := range t.Dst {
					s.Protocol(t.Src, a, t.Ptype)
				}
			}
			t.Launched = true
			return s
		}
	}
	return nil
}

func AddConnToShosetFromTopology(s *shoset.Shoset,Lname string, tt []*ShosetCreation) *shoset.Shoset {
	for _, t := range tt {
		if t.Lname == Lname {
			if t.Ptype == "pki" {
				s.InitPKI(t.Src)
			} else {
				for _, a := range t.Dst {
					s.Protocol(t.Src, a, t.Ptype)
				}
			}
			t.Launched = true
			return s
		}
	}
	return nil
}

func CreateShosetOnlyBindFromTopology(Lname string, tt []*ShosetCreation) *shoset.Shoset {
	for _, t := range tt {
		if !t.Launched && t.Lname == Lname {
			s := shoset.NewShoset(t.Lname, t.Stype)
			s.Protocol(t.Src, "", "")
			//s.Bind(t.Src)
			t.Launched = true
			return s
		}
	}
	return nil
}

func PrintManyShosets(s []*shoset.Shoset) {
	for i, t := range s {
		fmt.Println("\nShoset ", i, ": ", t)
	}
}

func WaitForManyShosets(s []*shoset.Shoset) {
	for _, t := range s {
		t.WaitForProtocols(10)
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
