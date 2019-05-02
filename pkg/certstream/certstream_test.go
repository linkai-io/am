package certstream_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/certstream"
)

func TestCertStream(t *testing.T) {
	//t.Skip("run off one")
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	userContext := amtest.CreateUserContext(1, 1)
	//ctx := context.Background()

	bd := &mock.BigDataService{}
	bd.AddCTSubdomainsFn = func(ctx context.Context, userContext am.UserContext, etld string, queryTime time.Time, subdomains map[string]*am.CTSubdomain) error {
		for _, v := range subdomains {
			t.Logf("added: %s %s", v.ETLD, v.Subdomain)
		}
		return nil
	}
	devModeQueryTime := time.Date(2019, time.February, 13, 0, 0, 0, 0, time.Local)

	bd.GetETLDsFn = func(ctx context.Context, userContext am.UserContext) ([]*am.CTETLD, error) {
		etlds := make([]*am.CTETLD, 1)
		etlds = append(etlds, &am.CTETLD{ETLD_ID: 1, ETLD: "cloudflaressl.com", QueryTimestamp: devModeQueryTime.UnixNano()})
		return etlds, nil
	}

	b := certstream.NewBatcher(userContext, bd, 5)
	b.Init()
	stream := certstream.New(b)
	stream.AddETLD("cloudflaressl.com")
	closeCh := make(chan struct{})
	if err := stream.Init(closeCh); err != nil {
		t.Fatalf("error init %v\n", err)
	}

	time.Sleep(10 * time.Second)
	close(closeCh)
}
