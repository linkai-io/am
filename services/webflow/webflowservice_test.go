package webflow_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/services/scangroup"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/address"
	"github.com/linkai-io/am/services/webflow"
)

var env string
var dbstring string

type OrgData struct {
	OrgName     string
	OrgID       int
	GroupName   string
	GroupID     int
	UserContext am.UserContext
	DB          *pgx.ConnPool
	WG          *sync.WaitGroup
	AddrClient  am.AddressService
	SGClient    am.ScanGroupService
}

func init() {
	var err error
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
	sec := secrets.NewSecretsCache(env, "")
	dbstring, err = sec.DBString(am.CustomWebFlowServiceKey)
	if err != nil {
		panic("error getting dbstring secret")
	}
}

func initOrg(orgName, groupName string, t *testing.T) (*webflow.Service, *OrgData) {
	wg := &sync.WaitGroup{}
	orgData := &OrgData{
		OrgName:   orgName,
		GroupName: groupName,
		WG:        wg,
	}
	sec := secrets.NewSecretsCache(env, "")
	addrdb, err := sec.DBString(am.AddressServiceKey)
	if err != nil {
		t.Fatal("error getting dbstring secret")
	}

	sgdb, err := sec.DBString(am.ScanGroupServiceKey)
	if err != nil {
		t.Fatal("error getting dbstring secret")
	}

	auth := amtest.MockEmptyAuthorizer()
	addrClient := address.New(auth)
	if err := addrClient.Init([]byte(addrdb)); err != nil {
		t.Fatalf("error inint addr: %#v\n", err)
	}
	orgData.AddrClient = addrClient

	sgClient := scangroup.New(auth)
	if err := sgClient.Init([]byte(sgdb)); err != nil {
		t.Fatalf("error inint sg: %#v\n", err)
	}
	orgData.SGClient = sgClient

	requester := &testRequester{t: t, wg: wg}

	service := webflow.New(auth, sgClient, addrClient, requester)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing webdata service: %s\n", err)
	}

	orgData.DB = amtest.InitDB(env, t)

	amtest.CreateOrg(orgData.DB, orgName, t)
	orgData.OrgID = amtest.GetOrgID(orgData.DB, orgName, t)
	orgData.GroupID = amtest.CreateScanGroup(orgData.DB, orgName, groupName, t)
	orgData.UserContext = amtest.CreateUserContext(orgData.OrgID, 1)

	return service, orgData
}

func TestNew(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	addrLen := 5
	addrs := amtest.GenerateAddrs(1, 1, addrLen)
	for i, a := range addrs {
		a.HostAddress = fmt.Sprintf("%d.example.com", i)
	}
	wg := &sync.WaitGroup{}
	wg.Add(addrLen)

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	addrClient := amtest.MockAddressService(1, addrs)
	sgClient := amtest.MockScanGroupService(1, 1)
	requester := &testRequester{t: t, wg: wg}

	service := webflow.New(auth, sgClient, addrClient, requester)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing webflow service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webflowcreate", "webflowcreate", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	addrLen := 100
	addrs := amtest.GenerateAddrs(org.OrgID, org.GroupID, addrLen)
	addrToAdd := make(map[string]*am.ScanGroupAddress, addrLen)
	for i, a := range addrs {
		a.HostAddress = fmt.Sprintf("%d.example.com", i)
		a.AddressHash = convert.HashAddress(a.IPAddress, a.HostAddress)
		addrToAdd[a.AddressHash] = a
	}

	if _, _, err := org.AddrClient.Update(ctx, org.UserContext, addrToAdd); err != nil {
		t.Fatalf("error adding addresses: %v\n", err)
	}

	org.WG.Add(addrLen)
	cfg := &am.CustomWebFlowConfig{
		OrgID:        org.OrgID,
		GroupID:      org.GroupID,
		WebFlowName:  "testwebflow",
		CreationTime: time.Now().UnixNano(),
		ModifiedTime: time.Now().UnixNano(),
		Deleted:      false,
		Configuration: &am.CustomRequestConfig{
			Method: "GET",
			URI:    "/admin",
			Headers: map[string]string{
				"blah": "xyz",
			},
			Body: "",
			Match: map[int32]string{
				0: "",
			},
			OnlyPort:   0,
			OnlyScheme: "",
		},
	}

	webFlowID, err := service.Create(ctx, org.UserContext, cfg)
	if err != nil {
		t.Fatalf("error creating web flow config: %#v\n", err)
	}

	if webFlowID == 0 {
		t.Fatalf("got back empty webflowid")
	}

	_, err = service.Start(ctx, org.UserContext, webFlowID)
	if err != nil {
		t.Fatalf("error starting web flow: %#v\n", err)
	}
	org.WG.Wait()
	f := &am.CustomWebFilter{
		OrgID:     org.OrgID,
		GroupID:   org.GroupID,
		WebFlowID: webFlowID,
		Filters:   &am.FilterType{},
		Start:     0,
		Limit:     1000,
	}

	_, results, err := service.GetResults(ctx, org.UserContext, f)
	for _, r := range results {
		t.Logf("%#v\n", r)
	}
}
