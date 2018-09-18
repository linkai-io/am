package e2e_test

import (
	"context"
	"flag"
	"fmt"
	"testing"

	balancerpb "github.com/bsm/grpclb/grpclb_balancer_v1"

	"github.com/linkai-io/am/am"
	"google.golang.org/grpc"
)

var enableTests bool

func init() {
	flag.BoolVar(&enableTests, "enable", false, "pass true to enable e2e tests")
}

func TestLoadBalancer(t *testing.T) {
	if !enableTests {
		return
	}
	targets := []string{am.OrganizationServiceKey, am.ScanGroupServiceKey,
		am.AddressServiceKey, am.DispatcherServiceKey, am.CoordinatorServiceKey, am.UserServiceKey}

	for _, target := range targets {
		cc, err := grpc.Dial(":8383", grpc.WithInsecure())
		if err != nil {
			t.Logf("error: %s\n", err)
			continue
		}

		bc := balancerpb.NewLoadBalancerClient(cc)
		resp, err := bc.Servers(context.Background(), &balancerpb.ServersRequest{
			Target: target,
		})

		if err != nil {
			t.Logf("error: %s\n", err)
			continue
		}

		for _, srv := range resp.Servers {
			fmt.Printf("target %s: %d\t%s\n", target, srv.Score, srv.Address)
		}
		cc.Close()
	}

}
