package portscanner_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/pkg/portscanner"
)

var testPorts = []int32{21, 22, 23, 25, 53, 80, 135, 139, 443, 445, 1443, 1723, 3306, 3389, 5432, 5900, 6379, 8000, 8080, 8443, 8500, 9500, 27017}

func TestScanIPv4(t *testing.T) {
	scan := portscanner.New()
	if err := scan.Init("ens33"); err != nil {
		t.Skip("must run as root :|")
		t.Fatalf("error initializing scanner: %v\n", err)
	}
	ctx := context.Background()
	targetIP := "209.126.252.34" //scanner1.linkai.io
	results, err := scan.ScanIPv4(ctx, targetIP, 10, testPorts)
	if err != nil {
		t.Fatalf("error scanning scanner: %v\n", err)
	}
	for _, open := range results.Open {
		t.Logf("open %d\n", open)
	}

	for _, closed := range results.Closed {
		t.Logf("closed: %d\n", closed)
	}
	t.Logf("in: %d open: %d closed: %d totals: %d", len(testPorts), len(results.Open), len(results.Closed), len(results.Open)+len(results.Closed))
}
