package utilsForTest

import (
	"context"
	"fmt"
	"time"

	"github.com/ditrit/shoset"
	"github.com/ditrit/shoset/msg"
)

func LoopUntilDone(tick time.Duration, ctx context.Context, callback func()) {
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

// Launche shoset using the list of instructions provided and add them to the list (the modified list id returned)
func CreateManyShosets(tt []*ShosetCreation, s []*shoset.Shoset, wait bool) []*shoset.Shoset {
	for i, t := range tt {
		if !t.Launched {
			s = append(s, shoset.NewShoset(t.Lname, t.ShosetType))
			if t.ProtocolType == "pki" {
				s[i].InitPKI(t.LocalAddress)
			} else {
				for _, a := range t.RemoteAddresses {
					s[i].Protocol(t.LocalAddress, a, t.ProtocolType)
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

// Create a single Shoset from the topology (for multiProcesses tests)
func CreateShosetFromTopology(Lname string, tt []*ShosetCreation) *shoset.Shoset {
	for _, t := range tt {
		if !t.Launched && t.Lname == Lname {
			s := shoset.NewShoset(t.Lname, t.ShosetType)
			if t.ProtocolType == "pki" {
				s.InitPKI(t.LocalAddress)
			} else {
				for _, a := range t.RemoteAddresses {
					s.Protocol(t.LocalAddress, a, t.ProtocolType)
				}
			}
			t.Launched = true
			return s
		}
	}
	return nil
}

// Same as CreateShosetFromTopology but only launches an empty protocol to relaunch the connections saved in the config file
func CreateShosetOnlyBindFromTopology(Lname string, tt []*ShosetCreation) *shoset.Shoset {
	for _, t := range tt {
		if !t.Launched && t.Lname == Lname {
			s := shoset.NewShoset(t.Lname, t.ShosetType)
			s.Protocol(t.LocalAddress, "", "")
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

// Waits for every shosets in the list to be ready
func WaitForManyShosets(s []*shoset.Shoset) {
	for _, t := range s {
		t.WaitForProtocols(10)
	}
}

// Forces reroute of every shosets in the list.
func RouteManyShosets(s []*shoset.Shoset, wait bool) {
	for _, t := range s {
		routing := msg.NewRoutingEvent(t.GetLogicalName(), true, 0, "")
		t.Send(routing)
	}
	if wait {
		time.Sleep(1 * time.Second)
	}
}
