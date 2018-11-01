package ns_test

import (
	"context"
	"os"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/ns"
)

const dnsServer = "1.1.1.1:53"
const localServer = "127.0.0.53:53"

func TestNS_Analyze(t *testing.T) {

	tests := []*am.ScanGroupAddress{
		&am.ScanGroupAddress{
			AddressID:    1,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "linkai.io",
			IPAddress:    "",
			DiscoveredBy: "input_list",
		},
		&am.ScanGroupAddress{
			AddressID:    2,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "",
			IPAddress:    "13.35.67.123",
			DiscoveredBy: "input_list",
		},
		&am.ScanGroupAddress{
			AddressID:    3,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "linkai.io",
			IPAddress:    "13.35.67.123",
			DiscoveredBy: "input_list",
		},
		&am.ScanGroupAddress{
			AddressID:    4,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "zonetransfer.me",
			IPAddress:    "",
			DiscoveredBy: "input_list",
		},
	}
	state := amtest.MockNSState()
	dc := dnsclient.New([]string{localServer}, 3)
	ns := ns.New(dc, state)
	ns.Init(nil)
	userContext := amtest.CreateUserContext(1, 1)
	ctx := context.Background()

	for _, tt := range tests {
		t.Logf("%d\n", tt.AddressID)
		ns.Analyze(ctx, userContext, tt)
	}
}

func TestNetflixInput(t *testing.T) {
	orgID := 1
	groupID := 1
	addrFile, err := os.Open("testdata/netflix.txt")
	if err != nil {
		t.Fatalf("error opening test data: %s\n", err)
	}

	addrs := amtest.AddrsFromInputFile(orgID, groupID, addrFile, t)

	state := amtest.MockNSState()
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 3)
	ns := ns.New(dc, state)
	ns.Init(nil)

	ctx := context.Background()
	userContext := amtest.CreateUserContext(orgID, 1)
	updated := &am.ScanGroupAddress{}
	netflix := &am.ScanGroupAddress{}
	for _, addr := range addrs {
		if addr.HostAddress == "www.netflix.com" {
			netflix = addr
			updated, _, err = ns.Analyze(ctx, userContext, addr)
			if err != nil {
				t.Fatalf("error analyzing: %s\n", err)
			}
		}
	}

	if updated.AddressID != netflix.AddressID {
		t.Fatalf("expected addr id: %d got %d\n", netflix.AddressID, updated.AddressID)
	}

	if updated.IPAddress == "" {
		t.Fatalf("did not get ip address for updated netflix\n")
	}
}
