package bigdata_test

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/bigdata"
)

var env string
var dbstring string

const serviceKey = "bigdataservice"

func init() {
	var err error
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
	sec := secrets.NewSecretsCache(env, "")
	dbstring, err = sec.DBString(serviceKey)
	if err != nil {
		panic("error getting dbstring secret")
	}
}

func TestNew(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := amtest.MockAuthorizer()
	service := bigdata.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing bigdata service: %s\n", err)
	}
}

func TestAddGetSubdomains(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := amtest.MockAuthorizer()
	service := bigdata.New(auth)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing bigdata service: %s\n", err)
	}
	now := time.Now()
	subdomains := amtest.BuildSubdomainsForCT("example.com", 3)

	if err := service.AddCTSubdomains(ctx, userContext, "example.com", now, subdomains); err != nil {
		t.Fatalf("error adding ct records: %#v\n", err)
	}

	defer func() {
		err := service.DeleteCTSubdomains(ctx, userContext, "example.com")
		if err != nil {
			t.Fatalf("failed to delete etld in ct subdomain tables: %#v\n", err)
		}
	}()

	addedTime, returned, err := service.GetCTSubdomains(ctx, userContext, "example.com")
	if err != nil {
		t.Fatalf("error getting CT subdomain records: %#v\n", err)
	}

	if addedTime.UnixNano() != now.UnixNano() {
		t.Fatalf("query time did not match insertion time: %v ~ %v\n", addedTime, now)
	}

	if len(subdomains) != len(returned) {
		t.Fatalf("%d did not match returned %d\n", len(subdomains), len(returned))
	}

	for sub := range subdomains {
		if _, ok := returned[sub]; !ok {
			t.Fatalf("subdomain in original input was ")
		}
	}

	// Add them again to ensure query timestamp is updated properly.
	now = time.Now()
	subdomains = amtest.BuildSubdomainsForCT("example.com", 6)
	if err := service.AddCTSubdomains(ctx, userContext, "example.com", now, subdomains); err != nil {
		t.Fatalf("error adding ct records: %#v\n", err)
	}

	addedTime, returned, err = service.GetCTSubdomains(ctx, userContext, "example.com")
	if err != nil {
		t.Fatalf("error getting CT subdomain records: %#v\n", err)
	}

	if addedTime.UnixNano() != now.UnixNano() {
		t.Fatalf("query time did not match insertion time: %v ~ %v\n", addedTime, now)
	}

	if len(subdomains) != len(returned) {
		t.Fatalf("%d did not match returned %d\n", len(subdomains), len(returned))
	}

	for sub := range subdomains {
		if _, ok := returned[sub]; !ok {
			t.Fatalf("subdomain in original input was ")
		}
	}
}

func TestAddGetCT(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := amtest.MockAuthorizer()
	service := bigdata.New(auth)
	ctx := context.Background()
	userContext := amtest.CreateUserContext(1, 1)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing bigdata service: %s\n", err)
	}
	now := time.Now()
	records := amtest.BuildCTRecords("example.com", now.UnixNano(), 3)

	if err := service.AddCT(ctx, userContext, "example.com", now, records); err != nil {
		t.Fatalf("error adding ct records: %#v\n", err)
	}

	defer func() {
		err := service.DeleteCT(ctx, userContext, "example.com")
		if err != nil {
			t.Fatalf("failed to delete etld in ct tables: %#v\n", err)
		}
	}()

	addedTime, returned, err := service.GetCT(ctx, userContext, "example.com")
	if err != nil {
		t.Fatalf("error getting CT records: %#v\n", err)
	}

	if addedTime.UnixNano() != now.UnixNano() {
		t.Fatalf("query time did not match insertion time: %v ~ %v\n", addedTime, now)
	}

	amtest.TestCompareCTRecords(records, returned, t)
}
