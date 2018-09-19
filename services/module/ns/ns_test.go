package ns_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/redisclient"
	"github.com/linkai-io/am/services/module/ns"
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
	state := &mock.NSState{}
	state.SubscribeFn = func(ctx context.Context, onStartFn redisclient.SubOnStart, onMessageFn redisclient.SubOnMessage, channels ...string) error {
		return nil
	}

	hosts := make(map[string]bool)
	state.DoNSRecordsFn = func(ctx context.Context, orgID int, scanGroupID int, expireSeconds int, zone string) (bool, error) {
		if _, ok := hosts[zone]; !ok {
			hosts[zone] = true
			return true, nil
		}
		return false, nil
	}
	ns := ns.New(state)
	ns.Init(nil)

	ctx := context.Background()

	for _, tt := range tests {
		t.Logf("%d\n", tt.AddressID)
		ns.Analyze(ctx, tt)
	}
}
