package web_test

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/mock"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/web"
	"github.com/linkai-io/am/services/webdata"
)

var dbstring string
var env string

type OrgData struct {
	OrgName     string
	OrgID       int
	GroupName   string
	GroupID     int
	UserContext am.UserContext
	DB          *pgx.ConnPool
}

func init() {
	var err error
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
	sec := secrets.NewSecretsCache(env, "")
	dbstring, err = sec.DBString(am.WebDataServiceKey)
	if err != nil {
		panic("error getting dbstring secret")
	}
}

func initOrg(orgName, groupName string, t *testing.T) (*webdata.Service, *OrgData) {
	orgData := &OrgData{
		OrgName:   orgName,
		GroupName: groupName,
	}

	auth := amtest.MockEmptyAuthorizer()

	service := webdata.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

	orgData.DB = amtest.InitDB(env, t)

	amtest.CreateOrg(orgData.DB, orgName, t)
	orgData.OrgID = amtest.GetOrgID(orgData.DB, orgName, t)
	orgData.GroupID = amtest.CreateScanGroup(orgData.DB, orgName, groupName, t)
	orgData.UserContext = amtest.CreateUserContext(orgData.OrgID, 1)

	return service, orgData
}

func TestWebInfraAnalyze(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}
	ctx := context.Background()

	browserPool := browser.NewGCDBrowserPool(2, amtest.MockWebDetector())
	if err := browserPool.Init(); err != nil {
		t.Fatalf("failed initializing browsers: %v\n", err)
	}
	defer browserPool.Close(ctx)
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 1)
	webDataClient, org := initOrg("testwebanalyze", "testwebanalyze", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)

	stater := amtest.MockWebState()
	storer := amtest.MockStorage()

	w := web.New(browserPool, webDataClient, dc, stater, storer)
	if err := w.Init(); err != nil {
		t.Fatalf("failed to init web module: %v\n", err)
	}

	_, _, err := w.Analyze(ctx, org.UserContext, address)
	if err != nil {
		t.Fatalf("error in analyze: %v\n", err)
	}

	filter := &am.WebResponseFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Start:   0,
		Limit:   1000,
		Filters: &am.FilterType{},
	}

	_, resps, err := webDataClient.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %v\n", err)
	}
	for _, resp := range resps {
		t.Logf("%#v\n", resp)
		if resp.AddressHash == "" {
			t.Fatalf("error address hash was empty")
		}
	}

	certFilter := &am.WebCertificateFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Start:   0,
		Limit:   1000,
		Filters: &am.FilterType{},
	}
	_, certs, err := webDataClient.GetCertificates(ctx, org.UserContext, certFilter)
	if err != nil {
		t.Fatalf("error getting responses: %v\n", err)
	}
	for _, cert := range certs {
		t.Logf("%#v\n", cert)
		if cert.AddressHash == "" {
			t.Fatalf("error address hash was empty")
		}
		if cert.IPAddress == "" {
			t.Fatalf("error address ip was empty")
		}
		if cert.Port == "" {
			t.Fatalf("error port was empty")
		}
	}

	snapFilter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Start:   0,
		Limit:   1000,
		Filters: &am.FilterType{},
	}
	_, snaps, err := webDataClient.GetSnapshots(ctx, org.UserContext, snapFilter)
	if err != nil {
		t.Fatalf("error getting responses: %v\n", err)
	}
	for _, snap := range snaps {
		t.Logf("%#v\n", snap)
		if snap.AddressHash == "" {
			t.Fatalf("error address hash was empty")
		}
		if snap.IPAddress == "" {
			t.Fatalf("error address ip was empty")
		}

		if snap.ResponsePort == 0 {
			t.Fatalf("error port was empty")
		}
	}
}

func TestWebAnalyze(t *testing.T) {
	ctx := context.Background()

	browserPool := browser.NewGCDBrowserPool(5, amtest.MockWebDetector())
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

func TestCoTAnalyze(t *testing.T) {
	ctx := context.Background()

	browserPool := browser.NewGCDBrowserPool(2, amtest.MockWebDetector())
	if err := browserPool.Init(); err != nil {
		t.Fatalf("failed initializing browsers: %v\n", err)
	}
	defer browserPool.Close(ctx)
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 1)

	stater := amtest.MockWebState()
	storer := amtest.MockStorage()

	webDataClient := &mock.WebDataService{}
	webDataClient.AddFn = func(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
		log.Info().Msgf("dom: %s", webData.SerializedDOM)
		ioutil.WriteFile("testdata/"+webData.SerializedDOMHash, []byte(webData.SerializedDOM), 0644)
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
		HostAddress:     "veracode.com",
		IPAddress:       "104.17.6.6",
		ConfidenceScore: 100,
		AddressHash:     convert.HashAddress("104.17.6.6", "veracode.com"),
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

func TestBannedIP(t *testing.T) {
	ctx := context.Background()

	browserPool := browser.NewGCDBrowserPool(2, amtest.MockWebDetector())
	if err := browserPool.Init(); err != nil {
		t.Fatalf("failed initializing browsers: %v\n", err)
	}
	defer browserPool.Close(ctx)
	dc := dnsclient.New([]string{"1.1.1.1:53"}, 1)

	stater := amtest.MockWebState()
	storer := amtest.MockStorage()

	webDataClient := &mock.WebDataService{}
	webDataClient.AddFn = func(ctx context.Context, userContext am.UserContext, webData *am.WebData) (int, error) {
		log.Info().Msgf("dom: %s", webData.SerializedDOM)
		ioutil.WriteFile("testdata/"+webData.SerializedDOMHash, []byte(webData.SerializedDOM), 0644)
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
		HostAddress:     "veracode.com",
		IPAddress:       "169.254.169.254",
		ConfidenceScore: 100,
		AddressHash:     convert.HashAddress("169.254.169.254", "veracode.com"),
	}

	_, newAddrs, err := w.Analyze(ctx, userContext, addr)
	if err != nil {
		t.Fatalf("failed to analyze example.com: %v\n", err)
	}
	if len(newAddrs) > 0 {
		t.Fatalf("should have had banned IP")
	}

	t.Logf("new addrs: %d\n", len(newAddrs))
	for _, v := range newAddrs {
		t.Logf("%#v\n", v)
	}

	addr = &am.ScanGroupAddress{
		OrgID:           1,
		GroupID:         1,
		HostAddress:     "veracode.com",
		IPAddress:       "10.0.1.1",
		ConfidenceScore: 100,
		AddressHash:     convert.HashAddress("10.0.1.1", "veracode.com"),
	}

	_, newAddrs, err = w.Analyze(ctx, userContext, addr)
	if err != nil {
		t.Fatalf("failed to analyze example.com: %v\n", err)
	}
	if len(newAddrs) > 0 {
		t.Fatalf("should have had banned IP")
	}
}
