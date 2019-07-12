package am_test

import (
	"testing"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/am"
)

func TestTCPPortChange(t *testing.T) {
	p := &am.Ports{
		Current:  &am.PortData{},
		Previous: &am.PortData{},
	}
	// test equal
	p.Current.TCPPorts = []int32{80, 443, 9090}
	p.Previous.TCPPorts = []int32{80, 443, 9090}
	open, closed, change := p.TCPChanges()
	if change == true {
		t.Fatalf("change detected when ports are equal")
	}

	if !amtest.SortEqualInt32(open, closed, t) {
		t.Fatalf("open and closed should be equal")
	}

	// test nil current
	p.Current.TCPPorts = nil
	p.Previous.TCPPorts = []int32{80, 443, 9090}
	open, closed, change = p.TCPChanges()
	if change == false {
		t.Fatalf("change should be false")
	}

	if !amtest.SortEqualInt32(closed, []int32{80, 443, 9090}, t) {
		t.Fatalf("closed ports and previous ports should be equal")
	}

	// test nil previous
	p.Current.TCPPorts = []int32{80, 443, 9090}
	p.Previous.TCPPorts = nil
	open, closed, change = p.TCPChanges()
	if change == false {
		t.Fatalf("change should be false")
	}

	if !amtest.SortEqualInt32(open, []int32{80, 443, 9090}, t) {
		t.Fatalf("closed ports and previous ports should be equal")
	}

	// test empty current
	p.Current.TCPPorts = []int32{}
	p.Previous.TCPPorts = []int32{80, 443, 9090}
	open, closed, change = p.TCPChanges()
	if change == false {
		t.Fatalf("change should be false")
	}

	if !amtest.SortEqualInt32(closed, []int32{80, 443, 9090}, t) {
		t.Fatalf("closed ports and previous ports should be equal")
	}

	// test empty previous
	p.Current.TCPPorts = []int32{80, 443, 9090}
	p.Previous.TCPPorts = []int32{}
	open, closed, change = p.TCPChanges()
	if change == false {
		t.Fatalf("change should be false")
	}

	if !amtest.SortEqualInt32(open, []int32{80, 443, 9090}, t) {
		t.Fatalf("closed ports and previous ports should be equal")
	}

	// test closed ports
	p.Current.TCPPorts = []int32{80, 443}
	p.Previous.TCPPorts = []int32{80, 443, 9090}
	open, closed, change = p.TCPChanges()
	if change == false {
		t.Fatalf("change should be false")
	}

	if !amtest.SortEqualInt32(closed, []int32{9090}, t) {
		t.Fatalf("closed ports should have 1 entry %v", closed)
	}

	// test open port
	p.Current.TCPPorts = []int32{80, 443, 9090}
	p.Previous.TCPPorts = []int32{80, 443}
	open, closed, change = p.TCPChanges()
	if change == false {
		t.Fatalf("change should be false")
	}

	if !amtest.SortEqualInt32(open, []int32{9090}, t) {
		t.Fatalf("open ports and should have 1 entry %v\n", open)
	}
}
