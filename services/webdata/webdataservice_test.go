package webdata_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/webdata"
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

func TestNew(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := webdata.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

}

func TestAdd(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdataadd", "webdataadd", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateWebData(org, address, "example.com", "93.184.216.34")

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	// test adding again
	_, err = service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}
}

func TestGetSnapshots(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetsnapshots", "webdatagetsnapshots", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateWebData(org, address, "example.com", "93.184.216.34")

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebSnapshotFilter{
		OrgID:             org.OrgID,
		GroupID:           org.GroupID,
		WithResponseTime:  false,
		SinceResponseTime: 0,
		Start:             0,
		Limit:             1000,
	}
	_, snapshots, err := service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting snapshots: %#v\n", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot got: %d\n", len(snapshots))
	}
}

func TestGetCertificates(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetcerts", "webdatagetcerts", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateWebData(org, address, "example.com", "93.184.216.34")

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebCertificateFilter{
		OrgID:             org.OrgID,
		GroupID:           org.GroupID,
		WithResponseTime:  false,
		SinceResponseTime: 0,
		Start:             0,
		Limit:             1000,
	}
	_, certs, err := service.GetCertificates(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting certs: %#v\n", err)
	}

	if len(certs) != 1 {
		t.Fatalf("expected 1 certs got: %d\n", len(certs))
	}
}

func TestGetResponses(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetresponses", "webdatagetresponses", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateWebData(org, address, "example.com", "93.184.216.34")

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebResponseFilter{
		OrgID:             org.OrgID,
		GroupID:           org.GroupID,
		WithResponseTime:  false,
		SinceResponseTime: 0,
		Start:             0,
		Limit:             1000,
	}

	_, certs, err := service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(certs) != 1 {
		t.Fatalf("expected 1 response got: %d\n", len(certs))
	}
}

func testCreateWebData(org *OrgData, address *am.ScanGroupAddress, host, ip string) *am.WebData {
	headers := make(map[string]string, 0)
	headers["host"] = host

	response := &am.HTTPResponse{
		Scheme:            "http",
		HostAddress:       host,
		IPAddress:         ip,
		ResponsePort:      "80",
		RequestedPort:     "80",
		Status:            200,
		StatusText:        "HTTP 200 OK",
		URL:               fmt.Sprintf("http://%s/", host),
		Headers:           headers,
		MimeType:          "text/html",
		RawBody:           "",
		RawBodyLink:       "s3://data/1/1/1/1",
		RawBodyHash:       "1111",
		ResponseTimestamp: time.Now().UnixNano(),
		IsDocument:        true,
		WebCertificate: &am.WebCertificate{
			Protocol:                          "h2",
			KeyExchange:                       "kex",
			KeyExchangeGroup:                  "keg",
			Cipher:                            "aes",
			Mac:                               "1234",
			CertificateValue:                  0,
			SubjectName:                       host,
			SanList:                           []string{"www." + host, host},
			ValidFrom:                         time.Now().UnixNano(),
			ValidTo:                           time.Now().UnixNano(),
			CertificateTransparencyCompliance: "unknown",
		},
	}
	responses := make([]*am.HTTPResponse, 1)
	responses[0] = response

	webData := &am.WebData{
		Address:           address,
		Responses:         responses,
		SerializedDOM:     "",
		SerializedDOMHash: "1234",
		SerializedDOMLink: "s3:/1/2/3/4",
		Snapshot:          "",
		SnapshotLink:      "s3://snapshot/1",
		ResponseTimestamp: time.Now().UnixNano(),
	}

	return webData
}
