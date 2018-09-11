package ns_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/modules/ns"
	"github.com/linkai-io/am/pkg/state/redis"
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
		&am.ScanGroupAddress{
			AddressID:    4,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "zonetransfer.me",
			IPAddress:    "",
			DiscoveredBy: "input_list",
		},
	}
	state := redis.New()
	if err := state.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis\n")
	}
	ns := ns.New(state)
	ns.Init(nil)
	ctx := context.Background()

	for _, tt := range tests {
		t.Logf("%d\n", tt.AddressID)
		ns.Analyze(ctx, tt)
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
