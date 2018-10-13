package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/services/module/web"
)

func TestGCDBrowser(t *testing.T) {
	b := web.NewGCDBrowser()
	defer b.Close()
	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	address := &am.ScanGroupAddress{
		HostAddress: "independent.co.uk",
		IPAddress:   "151.101.65.184",
	}

	d, r, err := b.Load(context.Background(), address, "http", "80")
	if err != nil {
		t.Fatalf("error during load: %v\n", err)
	}
	t.Logf("%s\n", d)
	for _, resp := range r {
		t.Logf("%v\n", resp)
	}
	t.Logf("sleeping...\n")
	time.Sleep(time.Second * 1)
}
