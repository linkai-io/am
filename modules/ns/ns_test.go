package ns_test

import (
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/modules/ns"
)

const dnsServer = "0.0.0.0:2053"
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
	}
	ns := ns.New(nil)
	ns.Init(nil)
	for _, tt := range tests {
		t.Logf("%d\n", tt.AddressID)
		ns.Analyze(tt)
	}
}

/*
addr: {
			am.ScanGroupAddress{
				OrgID:        1,
				GroupID:      1,
				HostAddress:  "",
				IPAddress:    "13.35.67.123",
				DiscoveredBy: "input_list",
			},
		},
*/
