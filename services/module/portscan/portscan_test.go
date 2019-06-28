package portscan_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/pkg/portscanner"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/portscan"
)

var testDNServer = []string{"1.1.1.1:53"}

func TestInit(t *testing.T) {
	dnsClient := dnsclient.New(testDNServer, 2)
	scanner := portscanner.NewLocalClient()
	module := portscan.New(scanner, dnsClient)
	if err := module.Init(nil); err != nil {
		t.Fatalf("error init module: %v\n", err)
	}

	group := amtest.CreateScanGroupOnly(1, 1)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	module.AddGroup(ctx, userContext, group)
	module.RemoveGroup(ctx, userContext, group.OrgID, group.GroupID)
}

func TestPortScan(t *testing.T) {
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	group := amtest.CreateScanGroupOnly(1, 1)

	dnsClient := dnsclient.New(testDNServer, 2)
	scanner := portscanner.NewLocalClient()
	module := portscan.New(scanner, dnsClient)
	if err := module.Init(nil); err != nil {
		t.Fatalf("error init module: %v\n", err)
	}

	addr := amtest.CreateAddressOnly(1, 1, "209.126.252.34", "scanner1.linkai.io", t)
	module.AddGroup(ctx, userContext, group)

	updatedAddr, ports, err := module.Analyze(ctx, userContext, addr)
	if err != nil {
		t.Fatalf("error analyzing ip %#v\n", err)
	}
	if ports == nil {
		t.Fatalf("ports was nil")
	}
	if ports.Ports.Current.IPAddress != "209.126.252.34" {
		t.Fatalf("expected IP Address to be set got %v\n", ports.Ports.Current.IPAddress)
	}
	if len(ports.Ports.Current.TCPPorts) != 1 {
		t.Fatalf("expected 1 port opened got %d\n", len(ports.Ports.Current.TCPPorts))
	}
	t.Logf("%#v\n", ports.Ports.Current)
	if updatedAddr.HostAddress != "scanner1.linkai.io" {
		t.Fatalf("error did not get valid host address back %#v", updatedAddr)
	}

}
