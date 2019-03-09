package webdata_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/convert"

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

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := webdata.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing webdata service: %s\n", err)
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

func TestOrgStats(t *testing.T) {
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
	oid, stats, err := service.OrgStats(ctx, org.UserContext)
	if err != nil {
		t.Fatalf("error getting stats: %v\n", err)
	}
	if oid != org.UserContext.GetOrgID() {
		t.Fatalf("mismatched org id")
	}

	if len(stats) != 1 {
		t.Logf("%#v\n", stats)
		t.Fatalf("expected one groups response, got %d\n", len(stats))
	}
	t.Logf("%#v\n", stats[0])
	if stats[0].ExpiringCerts15Days != 15 && stats[0].ExpiringCerts30Days != 30 {
		t.Fatalf("expected 15 expiring in 15, and 30 expiring in 30, got %d %d\n", stats[0].ExpiringCerts15Days, stats[0].ExpiringCerts30Days)
	}

	if stats[0].GroupID != org.GroupID && stats[0].OrgID != org.OrgID {
		t.Fatalf("org/group not set")
	}

	if stats[0].UniqueWebServers != 100 {
		t.Fatalf("expected 100 unique servers")
	}

	if stats[0].ServerTypes["Apache 1.0.55"] != 1 {
		t.Fatalf("expected only 1 for Apache 1.0.55, got %d\n", stats[0].ServerTypes["Apache 1.0.55"])
	}

	_, groupStats, err := service.GroupStats(ctx, org.UserContext, stats[0].GroupID)
	if err != nil {
		t.Fatalf("error getting group stats: %v\n", err)
	}

	if groupStats.UniqueWebServers != stats[0].UniqueWebServers {
		t.Fatalf("unique server did not match")
	}

	// add with nil server header
	web := testCreateWebData(org, address, "test.com", "1.1.1.1")
	if _, err := service.Add(ctx, org.UserContext, web); err != nil {
		t.Fatalf("error adding web data")
	}

	_, stats, err = service.OrgStats(ctx, org.UserContext)
	if err != nil {
		t.Fatalf("error getting stats with nil server header %v\n", err)
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
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	_, snapshots, err := service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting snapshots: %#v\n", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot got: %d\n", len(snapshots))
	}

	if snapshots[0].HostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if snapshots[0].IPAddress != "93.184.216.34" {
		t.Fatalf("expected address id ip address to be set")
	}

	if snapshots[0].URL != "http://example.com/" {
		t.Fatalf("expected URL: http://example.com/ got %s\n", snapshots[0].URL)
	}
	catLen := len(snapshots[0].TechCategories)
	nameLen := len(snapshots[0].TechNames)
	verLen := len(snapshots[0].TechVersions)
	matchLocLen := len(snapshots[0].TechMatchLocations)
	matchDataLen := len(snapshots[0].TechMatchData)
	iconLen := len(snapshots[0].TechIcons)
	webLen := len(snapshots[0].TechWebsites)
	avg := (catLen + nameLen + verLen + matchLocLen + matchDataLen + iconLen + webLen) / 7
	if avg != 3 {
		t.Fatalf("tech data lengths did not match")
	}
	t.Logf("%#v\n", snapshots[0])
}

func TestGetSnapshotsEmptyTech(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetsnapshotsemptytech", "webdatagetsnapshotsemptytech", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateWebData(org, address, "example.com", "93.184.216.34")
	webData.DetectedTech = nil

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	_, snapshots, err := service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting snapshots: %#v\n", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot got: %d\n", len(snapshots))
	}

	if snapshots[0].HostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if snapshots[0].IPAddress != "93.184.216.34" {
		t.Fatalf("expected address id ip address to be set")
	}

	if snapshots[0].URL != "http://example.com/" {
		t.Fatalf("expected URL: http://example.com/ got %s\n", snapshots[0].URL)
	}
	t.Logf("%#v\n", snapshots[0])

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
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
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
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}

	_, responses, err := service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(responses) != 1 {
		t.Fatalf("expected 1 response got: %d\n", len(responses))
	}

	if responses[0].HostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if responses[0].IPAddress != "93.184.216.34" {
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
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddInt64("after_request_time", 0)
	filter.Filters.AddString("header_names", "content-type")
	filter.Filters.AddString("not_header_names", "x-content-type")
	filter.Filters.AddString("mime_type", "text/html")

	_, responses, err := service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(responses) != 100 {
		t.Fatalf("expected 100 response got: %d\n", len(responses))
	}

	if responses[0].HostAddress != "example.com" {
		t.Fatalf("expected address id host address to be set")
	}

	if responses[0].IPAddress != "93.184.216.34" {
		t.Fatalf("expected address id ip address to be set")
	}

	if responses[0].URLRequestTimestamp == 0 {
		t.Fatalf("expected URLRequestTimestamp to be set (from webdata) got 0")
	}

	filter.Filters.AddBool("latest_only", true)
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

	// test sql injection
	filter = &am.WebResponseFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddString("header_names", "' or 1=1--")
	filter.Filters.AddString("not_header_names", "' or 1=1--")
	filter.Filters.AddString("mime_type", "' or 1=1--")
	filter.Filters.AddString("starts_host_address", "' or 1=1--")

	_, _, err = service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
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
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
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

	filter.Filters.AddBool("latest_only", true)
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
		headers["server"] = fmt.Sprintf("Apache 1.0.%d", i)
		headers["content-type"] = "text/html"

		response := &am.HTTPResponse{
			OrgID:               address.OrgID,
			GroupID:             address.GroupID,
			Scheme:              "http",
			AddressHash:         convert.HashAddress(ip, host),
			HostAddress:         host,
			IPAddress:           ip,
			ResponsePort:        "80",
			RequestedPort:       "80",
			Status:              200,
			StatusText:          "HTTP 200 OK",
			URL:                 fmt.Sprintf("http://%s/%d", host, urlIndex),
			Headers:             headers,
			MimeType:            "text/html",
			RawBody:             "",
			RawBodyLink:         "s3://data/1/1/1/1",
			RawBodyHash:         "1111",
			ResponseTimestamp:   time.Now().UnixNano(),
			URLRequestTimestamp: 0,
			IsDocument:          true,
			WebCertificate: &am.WebCertificate{
				ResponseTimestamp: time.Now().UnixNano(),
				HostAddress:       host,
				IPAddress:         ip,
				AddressHash:       convert.HashAddress(ip, host),
				Port:              "443",
				Protocol:          "h2",
				KeyExchange:       "kex",
				KeyExchangeGroup:  "keg",
				Cipher:            "aes",
				Mac:               "1234",
				CertificateValue:  0,
				SubjectName:       host,
				SanList: []string{
					"www." + insertHost,
					insertHost,
				},
				Issuer:                            "",
				ValidFrom:                         time.Now().Unix(),
				ValidTo:                           time.Now().Add(time.Hour * time.Duration(24*i)).Unix(),
				CertificateTransparencyCompliance: "unknown",
				IsDeleted:                         false,
			},
			IsDeleted: false,
		}
		responses = append(responses, response)
		urlIndex++

		if i%10 == 0 {
			groupIdx++
			data := &am.WebData{
				Address:             address,
				Responses:           responses,
				SnapshotLink:        "s3://snapshot/1",
				URL:                 fmt.Sprintf("http://%s/%d", host, urlIndex),
				Scheme:              "http",
				AddressHash:         convert.HashAddress(ip, host),
				HostAddress:         host,
				IPAddress:           ip,
				ResponsePort:        80,
				SerializedDOMHash:   "1234",
				SerializedDOMLink:   "s3:/1/2/3/4",
				ResponseTimestamp:   time.Now().UnixNano(),
				URLRequestTimestamp: time.Now().Add(time.Hour * -time.Duration(groupIdx*24)).UnixNano(),
				DetectedTech: map[string]*am.WebTech{"3dCart": &am.WebTech{
					Matched:  "1.1.11,1.1.11",
					Version:  "1.1.11",
					Location: "headers",
				},
					"jQuery": &am.WebTech{
						Matched:  "1.1.11,1.1.11",
						Version:  "1.1.11",
						Location: "script",
					},
				},
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
		Scheme:              "http",
		AddressHash:         convert.HashAddress(ip, host),
		HostAddress:         host,
		IPAddress:           ip,
		ResponsePort:        "80",
		RequestedPort:       "80",
		Status:              200,
		StatusText:          "HTTP 200 OK",
		URL:                 fmt.Sprintf("http://%s/", host),
		Headers:             headers,
		MimeType:            "text/html",
		RawBody:             "",
		RawBodyLink:         "s3://data/1/1/1/1",
		RawBodyHash:         "1111",
		ResponseTimestamp:   time.Now().UnixNano(),
		URLRequestTimestamp: 0,
		IsDocument:          true,
		WebCertificate: &am.WebCertificate{
			ResponseTimestamp: time.Now().UnixNano(),
			HostAddress:       host,
			IPAddress:         ip,
			AddressHash:       convert.HashAddress(ip, host),
			Port:              "443",
			Protocol:          "h2",
			KeyExchange:       "kex",
			KeyExchangeGroup:  "keg",
			Cipher:            "aes",
			Mac:               "1234",
			CertificateValue:  0,
			SubjectName:       host,
			SanList: []string{
				"www." + host,
				host,
			},
			Issuer:                            "",
			ValidFrom:                         time.Now().Unix(),
			ValidTo:                           time.Now().Add(time.Hour * time.Duration(24)).Unix(),
			CertificateTransparencyCompliance: "unknown",
			IsDeleted:                         false,
		},
		IsDeleted: false,
	}
	responses := make([]*am.HTTPResponse, 1)
	responses[0] = response

	webData := &am.WebData{
		Address:             address,
		Responses:           responses,
		Snapshot:            "",
		SnapshotLink:        "s3://snapshot/1",
		URL:                 fmt.Sprintf("http://%s/", host),
		Scheme:              "http",
		AddressHash:         convert.HashAddress(ip, host),
		HostAddress:         host,
		IPAddress:           ip,
		ResponsePort:        80,
		SerializedDOMHash:   "1234",
		SerializedDOMLink:   "s3:/1/2/3/4",
		ResponseTimestamp:   time.Now().UnixNano(),
		URLRequestTimestamp: time.Now().UnixNano(),
		DetectedTech: map[string]*am.WebTech{"3dCart": &am.WebTech{
			Matched:  "1.1.11,1.1.11",
			Version:  "1.1.11",
			Location: "headers",
		},
			"jQuery": &am.WebTech{
				Matched:  "1.1.11,1.1.11",
				Version:  "1.1.11",
				Location: "script",
			},
		},
	}

	return webData
}

/*
func TestPopulateWeb(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("populatetest", "populatetest", t)
	//defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := testCreateMultiWebData(org, address, "example.com", "93.184.216.34")

	for i, web := range webData {
		t.Logf("%d: %d\n", i, len(web.Responses))
		_, err := service.Add(ctx, org.UserContext, web)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}
}
*/
