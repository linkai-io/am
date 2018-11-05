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
	sec := secrets.NewDBSecrets(env, "")
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
