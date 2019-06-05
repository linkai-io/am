package webdata_test

import (
	"context"
	"flag"
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
		t.Fatalf("error initalizing webdata service: %s\n", err)
	}

	orgData.DB = amtest.InitDB(env, t)
	amtest.DeleteOrg(orgData.DB, orgName, t)
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
	webData := amtest.CreateWebData(address, "example.com", "93.184.216.34")

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
	service, org := initOrg("webdataorgstats", "webdataorgstats", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateMultiWebData(address, "example.com", "93.184.216.34")

	for _, web := range webData {
		_, err := service.Add(ctx, org.UserContext, web)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}
	amtest.RunAggregates(org.DB, t)
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
	// we only expect 60 not 100 because we do wb.url_request_time > now()-'7 days'::interval
	// so some of the responses are not included
	if stats[0].UniqueWebServers != 60 {
		t.Fatalf("expected 60 unique servers got: %d\n", stats[0].UniqueWebServers)
	}

	for i, serverType := range stats[0].ServerTypes {
		if serverType == "Apache 1.0.55" && stats[0].ServerCounts[i] != 1 {
			t.Fatalf("expected only 1 for Apache 1.0.55, got %d\n", stats[0].ServerCounts[i])
		}
	}

	_, groupStats, err := service.GroupStats(ctx, org.UserContext, stats[0].GroupID)
	if err != nil {
		t.Fatalf("error getting group stats: %v\n", err)
	}

	if groupStats.UniqueWebServers != stats[0].UniqueWebServers {
		t.Fatalf("unique server did not match")
	}

	// add with nil server header
	web := amtest.CreateWebData(address, "test.com", "1.1.1.1")
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
	webData := amtest.CreateWebData(address, "example.com", "93.184.216.34")

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   time.Now().UnixNano(),
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

	if snapshots[0].LoadURL != "http://example.com/" {
		t.Fatalf("expected LoadURL: http://example.com/ got %s\n", snapshots[0].LoadURL)
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

func TestGetMultiSnapshots(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetsnapshots", "webdatagetsnapshots", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateMultiWebData(address, "example.com", "93.184.216.34")

	for _, data := range webData {
		_, err := service.Add(ctx, org.UserContext, data)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}

	ids := make(map[int64]struct{})
	lastIndex := time.Now().UnixNano()
	for {
		filter := &am.WebSnapshotFilter{
			OrgID:   org.OrgID,
			GroupID: org.GroupID,
			Filters: &am.FilterType{},
			Start:   lastIndex,
			Limit:   1,
		}
		t.Logf("querying with %d\n", lastIndex)
		_, snapshots, err := service.GetSnapshots(ctx, org.UserContext, filter)
		if err != nil {
			t.Fatalf("error getting snapshots: %#v\n", err)
		}

		if snapshots == nil || len(snapshots) == 0 {
			break
		}

		t.Logf("len: %d\n", len(snapshots))
		for _, s := range snapshots {
			t.Logf("%d %d %d\n", s.SnapshotID, lastIndex, s.URLRequestTimestamp)
			if s.URLRequestTimestamp < lastIndex {
				lastIndex = s.URLRequestTimestamp
			}

			if _, exists := ids[s.SnapshotID]; exists {
				t.Fatalf("error got duplicate snapshot id during filter")
			}
			ids[s.SnapshotID] = struct{}{}
		}
	}
	if len(ids) != 10 {
		t.Fatalf("expected 10 ids got %d\n", len(ids))
	}
}

func TestGetSnapshotsEmptyTech(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetsnapshotsemptytech", "webdatagetsnapshotsemptytech", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateWebData(address, "example.com", "93.184.216.34")
	webData.DetectedTech = nil

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   time.Now().UnixNano(),
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
	if snapshots[0].LoadURL != "http://example.com/" {
		t.Fatalf("expected LoadURL: http://example.com/ got %s\n", snapshots[0].LoadURL)
	}
	t.Logf("%#v\n", snapshots[0])
}

func TestGetSnapshotsWithTech(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetsnapshotswithtech", "webdatagetsnapshotswithtech", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateWebData(address, "example.com", "93.184.216.34")
	//webData.DetectedTech = nil

	_, err := service.Add(ctx, org.UserContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	filter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   time.Now().UnixNano(),
		Limit:   1000,
	}

	filter.Filters.AddString(am.FilterWebTechType, "jquery")
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
	if snapshots[0].LoadURL != "http://example.com/" {
		t.Fatalf("expected LoadURL: http://example.com/ got %s\n", snapshots[0].LoadURL)
	}
	// while it seems like this should be 2, 3dCart is returned 2x, one for CMS, one for ECommerce
	if len(snapshots[0].TechNames) != 3 {
		t.Fatalf("expected three tech names got %d\n", len(snapshots[0].TechNames))
	}

	found := false
	for _, tech := range snapshots[0].TechNames {
		if tech == "jQuery" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("failed to find jquery in results")
	}
	t.Logf("%#v\n", snapshots[0])

	filter.Filters.AddString(am.FilterWebTechTypeVersion, "1.1.11")
	_, snapshots, err = service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting snapshots: %#v\n", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("expected 1 snapshot got: %d\n", len(snapshots))
	}
	if len(snapshots[0].TechNames) != 3 {
		t.Fatalf("expected three tech names got %d\n", len(snapshots[0].TechNames))
	}

	found = false
	for _, tech := range snapshots[0].TechNames {
		if tech == "jQuery" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("failed to find jquery in results")
	}

	// try not match
	filter.Filters = &am.FilterType{}
	filter.Filters.AddString(am.FilterWebTechType, "jquery")
	filter.Filters.AddString(am.FilterWebTechTypeVersion, "1.1.12")
	_, snapshots, err = service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting snapshots: %#v\n", err)
	}

	if len(snapshots) != 0 {
		t.Fatalf("expected 0 snapshot got: %d\n", len(snapshots))
	}
}

func TestGetSnapshotsWithDomain(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webdatagetsnapshotswithdomain", "webdatagetsnapshotswithdomain", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateMultiWebData(address, "example.com", "93.184.216.34")

	for _, data := range webData {
		_, err := service.Add(ctx, org.UserContext, data)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}

	filter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   time.Now().UnixNano(),
		Limit:   1000,
	}

	filter.Filters.AddString(am.FilterWebDependentHostAddress, "example.com")
	_, snapshots, err := service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting snapshots: %#v\n", err)
	}

	if len(snapshots) != 10 {
		t.Fatalf("expected 1 snapshot got: %d\n", len(snapshots))
	}

	// while it seems like this should be 2, 3dCart is returned 2x, one for CMS, one for ECommerce
	if len(snapshots[0].TechNames) != 3 {
		t.Logf("%#v\n", snapshots[0])
		for _, tech := range snapshots[0].TechNames {
			t.Logf("%s\n", tech)
		}
		t.Fatalf("expected three tech names got %d\n", len(snapshots[0].TechNames))
	}

	found := false
	for _, tech := range snapshots[0].TechNames {
		if tech == "jQuery" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("failed to find jquery in results")
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
	webData := amtest.CreateWebData(address, "example.com", "93.184.216.34")

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
	webData := amtest.CreateWebData(address, "example.com", "93.184.216.34")

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
	webData := amtest.CreateMultiWebData(address, "example.com", "93.184.216.34")

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
	filter.Filters.AddInt64(am.FilterWebAfterURLRequestTime, 0)
	filter.Filters.AddString(am.FilterWebHeaderNames, "content-type")
	filter.Filters.AddString(am.FilterWebNotHeaderNames, "x-content-type")
	filter.Filters.AddString(am.FilterWebMimeType, "text/html")

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

	filter.Filters.AddBool(am.FilterWebLatestOnly, true)
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
	filter.Filters.AddString(am.FilterWebHeaderNames, "' or 1=1--")
	filter.Filters.AddString(am.FilterWebNotHeaderNames, "' or 1=1--")
	filter.Filters.AddString(am.FilterWebMimeType, "' or 1=1--")
	filter.Filters.AddString(am.FilterWebStartsHostAddress, "' or 1=1--")

	_, _, err = service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	// test server type filter
	filter = &am.WebResponseFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddString(am.FilterWebEqualsServerType, "Apache 1.0.1")

	_, resp, err := service.GetResponses(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	for _, rp := range resp {
		t.Logf("%#v", rp)
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
	webData := amtest.CreateMultiWebData(address, "example.com", "93.184.216.34")

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

	filter.Filters.AddBool(am.FilterWebLatestOnly, true)
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

	// test url list single query via url_request_timestamp
	filter = &am.WebResponseFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddInt64(am.FilterWebEqualsURLRequestTime, urlLists[0].URLRequestTimestamp)
	oid, urlLists, err = service.GetURLList(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}

	if len(urlLists) != 1 {
		t.Fatalf("expected 1 urllists got: %d\n", len(urlLists))
	}

	// test after request time
	filter = &am.WebResponseFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddInt64(am.FilterWebAfterURLRequestTime, time.Now().Add(time.Hour*-72).UnixNano())
	oid, urlLists, err = service.GetURLList(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}
	if len(urlLists) != 2 {
		t.Fatalf("expected 2 urllists got: %d\n", len(urlLists))
	}

}

func TestGetDomainDependency(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("domaindep", "domaindep", t)
	defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateMultiWebDataWithSub(address, "example.com", "93.184.216.34", 100)

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
	filter.Filters.AddInt64(am.FilterWebAfterURLRequestTime, time.Now().Add(time.Hour*-(7*24)).UnixNano())
	oid, domains, err := service.GetDomainDependency(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}

	if oid != org.OrgID {
		t.Fatalf("oid %v did not equal orgID: %v\n", oid, org.OrgID)
	}
	for _, d := range domains.Links {
		t.Logf("%#v\n", d)
	}

	for _, d := range domains.Nodes {
		t.Logf("%#v\n", d)
	}
}

func TestArchive(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("webarchive", "webarchive", t)
	//defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateMultiWebDataWithSub(address, "example.com", "93.184.216.34", 100)

	for i, web := range webData {
		t.Logf("%d: %d\n", i, len(web.Responses))
		_, err := service.Add(ctx, org.UserContext, web)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}

	filter := &am.WebSnapshotFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddInt64(am.FilterWebBeforeURLRequestTime, time.Now().Add(time.Hour*-(5*24)).UnixNano())
	_, snapshots, err := service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}
	snapLen := len(snapshots)
	t.Logf("snapshot len: %d\n", snapLen)
	if snapLen != 6 {
		t.Fatalf("expected 6 snapshots to archive, got %d\n", snapLen)
	}

	respFilter := &am.WebResponseFilter{
		OrgID:   org.OrgID,
		GroupID: org.GroupID,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	respFilter.Filters.AddInt64(am.FilterWebBeforeURLRequestTime, time.Now().Add(time.Hour*-(5*24)).UnixNano())
	_, responses, err := service.GetResponses(ctx, org.UserContext, respFilter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	respLen := len(responses)
	if respLen != 60 {
		t.Fatalf("expecetd 60 responses to archive, got %d\n", snapLen)
	}

	group := &am.ScanGroup{
		OrgID:                org.OrgID,
		GroupID:              org.GroupID,
		GroupName:            org.GroupName,
		CreationTime:         time.Now().UnixNano(),
		CreatedBy:            "test",
		CreatedByID:          org.UserContext.GetUserID(),
		ModifiedBy:           "test",
		ModifiedByID:         org.UserContext.GetUserID(),
		ModifiedTime:         time.Now().UnixNano(),
		OriginalInputS3URL:   "s3://empty",
		ModuleConfigurations: amtest.CreateModuleConfig(),
		Paused:               false,
		Deleted:              false,
		LastPausedTime:       0,
		ArchiveAfterDays:     5,
	}

	if _, _, err := service.Archive(ctx, org.UserContext, group, time.Now()); err != nil {
		t.Fatalf("error during archival: %#v\n", err)
	}

	_, snapshots, err = service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected 0 snapshots, got %d\n", len(snapshots))
	}

	_, responses, err = service.GetResponses(ctx, org.UserContext, respFilter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(responses) != 0 {
		t.Fatalf("expected 0 responses, got %d\n", len(responses))
	}

	filter.Filters = &am.FilterType{}
	_, snapshots, err = service.GetSnapshots(ctx, org.UserContext, filter)
	if err != nil {
		t.Fatalf("error getting url list: %v\n", err)
	}
	if len(snapshots) != 4 {
		t.Fatalf("expected 4 snapshots to not be archived, got %d\n", len(snapshots))
	}

	respFilter.Filters = &am.FilterType{}
	_, responses, err = service.GetResponses(ctx, org.UserContext, respFilter)
	if err != nil {
		t.Fatalf("error getting responses: %#v\n", err)
	}

	if len(responses) != 40 {
		t.Fatalf("expected 40 responses to not be archived, got %d\n", len(responses))
	}
}

func TestDeletePopulateWeb(t *testing.T) {
	t.Skip("uncomment to populate data")
	db := amtest.InitDB(env, t)
	amtest.DeleteOrg(db, "populatetest", t)

}
func TestPopulateWeb(t *testing.T) {
	t.Skip("uncomment to populate data")
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	service, org := initOrg("populatetest", "populatetest", t)
	//defer amtest.DeleteOrg(org.DB, org.OrgName, t)

	address := amtest.CreateScanGroupAddress(org.DB, org.OrgID, org.GroupID, t)
	webData := amtest.CreateMultiWebDataWithSub(address, "example.com", "93.184.216.34", 20000)

	for i, web := range webData {
		t.Logf("%d: %d\n", i, len(web.Responses))
		_, err := service.Add(ctx, org.UserContext, web)
		if err != nil {
			t.Fatalf("failed: %v\n", err)
		}
	}
}
