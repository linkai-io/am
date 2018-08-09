package address_test

import (
	"context"
	"flag"
	"fmt"
	"testing"
	"time"

	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/amtest"
	"gopkg.linkai.io/v1/repos/am/pkg/secrets"
	"gopkg.linkai.io/v1/repos/am/services/address"

	"gopkg.linkai.io/v1/repos/am/mock"
)

var env string
var dbstring string

const serviceKey = "addressservice"

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
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := address.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}
}

func TestAdd(t *testing.T) {
	ctx := context.Background()

	orgName := "addaddress"
	groupName := "addaddressgroup"

	auth := amtest.MockEmptyAuthorizer()

	service := address.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)

	groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
	userContext := amtest.CreateUserContext(orgID, 1)

	// test empty count first
	_, count, err := service.AddressCount(ctx, userContext, groupID)
	if err != nil {
		t.Fatalf("error getting empty count: %s\n", err)
	}

	if count != 0 {
		t.Fatalf("count should be zero for empty scangroup got: %d\n", count)
	}

	addresses := make([]*am.ScanGroupAddress, 0)
	now := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		a := &am.ScanGroupAddress{
			OrgID:           orgID,
			GroupID:         groupID,
			HostAddress:     "",
			IPAddress:       fmt.Sprintf("192.168.1.%d", i),
			DiscoveryTime:   now,
			DiscoveredBy:    "input_list",
			LastSeenTime:    0,
			IsSOA:           false,
			IsWildcardZone:  false,
			IsHostedService: false,
			Ignored:         false,
		}
		addresses = append(addresses, a)
	}

	_, count, err = service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addreses: %s\n", err)
	}
	if count != 100 {
		t.Fatalf("error expected count to be 100, got: %d\n", count)
	}

}
