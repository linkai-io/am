package portscanner_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/portscanner"
)

func TestServiceWithClient(t *testing.T) {
	serv := portscanner.NewService()
	if err := serv.Init(nil); err != nil {
		t.Skip("must run as root :|")
		t.Fatalf("error building server: %v\n", err)
	}

	go func() {
		if err := serv.Serve(); err != nil && err != http.ErrServerClosed {
			t.Fatalf("error calling serv %v\n", err)
		}
	}()
	defer serv.Shutdown()

	time.Sleep(100 * time.Millisecond) // give time for socket creation

	sclient := portscanner.NewSocketClient()
	if err := sclient.Init(nil); err != nil {
		t.Fatalf("error building client %v\n", err)
	}

	ctx := context.Background()
	res, err := sclient.PortScan(ctx, testScanner1IP, 10, testPorts)
	if err != nil {
		t.Fatalf("error running port scan %v\n", err)
	}

	if len(res.Open) != 1 {
		t.Fatalf("expected one port open got %d\n", len(res.Open))
	}

	if res.Open[0] != 22 {
		t.Fatalf("expected ssh port open got %d\n", res.Open[0])
	}

}
