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

	if snapshots[0].AddressIDHostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if snapshots[0].AddressIDIPAddress != "93.184.216.34" {
		t.Fatalf("expected address id ip address to be set")
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
		WithResponseTime:  true,
		SinceResponseTime: 0,
		Start:             0,
		Limit:             1000,
	}

	_, responses, err := service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(responses) != 1 {
		t.Fatalf("expected 1 response got: %d\n", len(responses))
	}

	if responses[0].AddressIDHostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if responses[0].AddressIDIPAddress != "93.184.216.34" {
		t.Fatalf("expected address id ip address to be set")
	}

	if responses[0].URLRequestTimestamp == 0 {
		t.Fatalf("expected URLRequestTimestamp to be set (from webdata) got 0")
	}
}

func TestGetResponsesWithAdvancedFilters(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetresponsesadvfilter", "webdatagetresponsesadvfilter", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateMultiWebData(org, address, "example.com", "93.184.216.34")

	for _, web := range webData {
		_, err := service.Add(ctx, org.UserContext, web)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}

	filter := &am.WebResponseFilter{
		OrgID:             org.OrgID,
		GroupID:           org.GroupID,
		WithResponseTime:  true,
		SinceResponseTime: 0,
		WithHeader:        "content-type",
		WithoutHeader:     "x-content-type",
		MimeType:          "text/html",
		Start:             0,
		Limit:             1000,
	}

	_, responses, err := service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(responses) != 100 {
		t.Fatalf("expected 100 response got: %d\n", len(responses))
	}

	if responses[0].AddressIDHostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if responses[0].AddressIDIPAddress != "93.184.216.34" {
		t.Fatalf("expected address id ip address to be set")
	}

	if responses[0].URLRequestTimestamp == 0 {
		t.Fatalf("expected URLRequestTimestamp to be set (from webdata) got 0")
	}

	filter.LatestOnlyValue = true

	_, responses, err = service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses again: %#v\n", err)
	}

	if len(responses) != 10 {
		t.Fatalf("expected 10 responses with latest got: %d\n", len(responses))
	}
	responseDay := time.Unix(0, responses[0].URLRequestTimestamp).Day()
	if responseDay != time.Now().Day()-1 {
		t.Fatalf("expected day to be now -1, got %d %d\n", responseDay, time.Now().Day()-1)
	}

	if responseDay != time.Unix(0, responses[9].URLRequestTimestamp).Day() {
		t.Fatalf("expected latest only, got %v and %v\n", time.Unix(0, responses[0].URLRequestTimestamp), time.Unix(0, responses[9].URLRequestTimestamp))
	}
}

func TestGetURLList(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatageturllist", "webdatageturllist", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateMultiWebData(org, address, "example.com", "93.184.216.34")

	for i, web := range webData {
		t.Logf("%d: %d\n", i, len(web.Responses))
		_, err := service.Add(ctx, org.UserContext, web)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}

	filter := &am.WebResponseFilter{
		OrgID:             org.OrgID,
		GroupID:           org.GroupID,
		WithResponseTime:  true,
		SinceResponseTime: 0,
		Start:             0,
		Limit:             1000,
	}
	oid, urlLists, err := service.GetURLList(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}

	if oid != org.OrgID {
		t.Fatalf("oid %v did not equal orgID: %v\n", oid, org.OrgID)
	}

	if len(urlLists) != 10 {
		t.Logf("%#v\n", urlLists[0])
		t.Fatalf("expected 10 rows of results, got %d\n", len(urlLists))
	}

	for i, urlList := range urlLists {
		if urlList.URLs == nil {
			t.Fatalf("error urls were empty")
		}
		// first iteration only has a single url
		if i > 1 && len(urlList.URLs) != 10 {
			t.Fatalf("expected 10 urls got: %d %#v\n", len(urlList.URLs), urlList.URLs)
		}
	}

	filter.LatestOnlyValue = true
	oid, urlLists, err = service.GetURLList(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}
	t.Logf("%#v\n", urlLists)
	if oid != org.OrgID {
		t.Fatalf("oid %v did not equal orgID: %v\n", oid, org.OrgID)
	}

	if len(urlLists) != 1 {
		t.Fatalf("expected 1 row of results, got %d\n", len(urlLists))
	}

	if len(urlLists[0].URLs) != 10 {
		t.Fatalf("expected 10 urls, got %d\n", len(urlLists[0].URLs))
	}
	requestDay := time.Unix(0, urlLists[0].URLRequestTimestamp).Day()
	if requestDay != time.Now().Day()-1 {
		t.Fatalf("last series of URLs should all have request timestamp of day - 1 got %d %d", requestDay, time.Now().Day()-1)
	}
}

func testCreateMultiWebData(org *OrgData, address *am.ScanGroupAddress, host, ip string) []*am.WebData {
	webData := make([]*am.WebData, 0)
	insertHost := host

	responses := make([]*am.HTTPResponse, 0)
	urlIndex := 0
	groupIdx := 0

	for i := 1; i < 101; i++ {
		headers := make(map[string]string, 0)
		headers["host"] = host
		headers["content-type"] = "text/html"

		response := &am.HTTPResponse{
			Scheme:            "http",
			HostAddress:       host,
			IPAddress:         ip,
			ResponsePort:      "80",
			RequestedPort:     "80",
			Status:            200,
			StatusText:        "HTTP 200 OK",
			URL:               fmt.Sprintf("http://%s/%d", host, urlIndex),
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
				SanList:                           []string{"www." + insertHost, insertHost},
				ValidFrom:                         time.Now().UnixNano(),
				ValidTo:                           time.Now().UnixNano(),
				CertificateTransparencyCompliance: "unknown",
			},
		}
		responses = append(responses, response)
		urlIndex++

		if i%10 == 0 {
			groupIdx++
			data := &am.WebData{
				Address:             address,
				Responses:           responses,
				SerializedDOM:       "",
				SerializedDOMHash:   "1234",
				SerializedDOMLink:   "s3:/1/2/3/4",
				Snapshot:            "",
				SnapshotLink:        "s3://snapshot/1",
				URLRequestTimestamp: time.Now().Add(time.Hour * -time.Duration(groupIdx*24)).UnixNano(),
				ResponseTimestamp:   time.Now().UnixNano(),
			}
			urlIndex = 0
			webData = append(webData, data)

			insertHost = fmt.Sprintf("%d.%s", i, host)
			responses = make([]*am.HTTPResponse, 0)
		}
	}

	return webData
}

func testCreateWebData(org *OrgData, address *am.ScanGroupAddress, host, ip string) *am.WebData {
	headers := make(map[string]string, 0)
	headers["host"] = host
	headers["content-type"] = "text/html"

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
		Address:             address,
		Responses:           responses,
		SerializedDOM:       "",
		SerializedDOMHash:   "1234",
		SerializedDOMLink:   "s3:/1/2/3/4",
		Snapshot:            "",
		SnapshotLink:        "s3://snapshot/1",
		URLRequestTimestamp: time.Now().UnixNano(),
		ResponseTimestamp:   time.Now().UnixNano(),
	}

	return webData
}
