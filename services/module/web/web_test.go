package web_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/mock"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/web"
)

func TestWebAnalyze(t *testing.T) {
	ctx := context.Background()

	browserPool := browser.NewGCDBrowserPool(5)
	if err := browserPool.Init(); err != nil {
		t.Fatalf("failed initializing browsers: %v\n", err)
	}
	defer browserPool.Close(ctx)
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 1)

	stater := amtest.MockWebState()
	storer := amtest.MockStorage()
	webDataClient := &mock.WebDataService{}
	webDataClient.AddFn = func(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
		return 1, nil
	}

	w := web.New(browserPool, webDataClient, dc, stater, storer)
	if err := w.Init(); err != nil {
		t.Fatalf("failed to init web module: %v\n", err)
	}

	userContext := amtest.CreateUserContext(1, 1)
	addr := &am.ScanGroupAddress{
		OrgID:           1,
		GroupID:         1,
		HostAddress:     "example.com",
		IPAddress:       "93.184.216.34",
		ConfidenceScore: 100,
		AddressHash:     convert.HashAddress("93.184.216.34", "example.com"),
	}

	_, newAddrs, err := w.Analyze(ctx, userContext, addr)
	if err != nil {
		t.Fatalf("failed to analyze example.com: %v\n", err)
	}

	t.Logf("new addrs: %d\n", len(newAddrs))
	for _, v := range newAddrs {
		t.Logf("%#v\n", v)
	}
}
