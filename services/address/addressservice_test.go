package address_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/address"

	"github.com/linkai-io/am/mock"
)

var env string
var dbstring string

const serviceKey = "addressservice"

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

func TestBuildGetFilter(t *testing.T) {
	service := address.New(nil)

	userContext := &am.UserContextData{OrgID: 1}
	filter := &am.ScanGroupAddressFilter{
		OrgID:                     1,
		GroupID:                   1,
		Start:                     0,
		Limit:                     1000,
		WithIgnored:               false,
		IgnoredValue:              false,
		WithBeforeLastScannedTime: false,
		WithAfterLastScannedTime:  false,
		AfterScannedTime:          0,
		BeforeScannedTime:         0,
		WithBeforeLastSeenTime:    false,
		WithAfterLastSeenTime:     false,
		AfterSeenTime:             0,
		BeforeSeenTime:            0,
		WithIsWildcard:            false,
		IsWildcardValue:           false,
		WithIsHostedService:       false,
		IsHostedServiceValue:      false,
		MatchesHost:               "",
		MatchesIP:                 "",
		NSRecord:                  0,
	}

	noFilters, noFilterArgs := service.BuildGetFilterQuery(userContext, filter)
	if len(noFilterArgs) != 4 {
		t.Fatalf("expected args len of 4, got %v\n", len(noFilterArgs))
	}

	if !strings.HasSuffix(noFilters, "$4") {
		t.Fatalf("query should have ended with $4 got %v\n", noFilters)
	}

	if noFilterArgs[0] != 1 && noFilterArgs[1] != 1 && noFilterArgs[2] != 0 && noFilterArgs[3] != 1000 {
		t.Fatalf("expected 1, 1, 0, 1000 for args, got %#v\n", noFilterArgs)
	}

	filter = &am.ScanGroupAddressFilter{
		OrgID:                     1,
		GroupID:                   1,
		Start:                     1000,
		Limit:                     2000,
		WithIgnored:               true,
		IgnoredValue:              false,
		WithBeforeLastScannedTime: true,
		WithAfterLastScannedTime:  true,
		AfterScannedTime:          0,
		BeforeScannedTime:         0,
		WithBeforeLastSeenTime:    true,
		WithAfterLastSeenTime:     true,
		AfterSeenTime:             0,
		BeforeSeenTime:            0,
		WithIsWildcard:            true,
		IsWildcardValue:           true,
		WithIsHostedService:       true,
		IsHostedServiceValue:      true,
		MatchesHost:               "asdf.com",
		MatchesIP:                 "192.168.9",
		NSRecord:                  33,
	}

	filters, filterArgs := service.BuildGetFilterQuery(userContext, filter)
	t.Logf("%s %#v\n", filters, filterArgs)
	if len(filterArgs) != 14 {
		t.Fatalf("expected args len of 14, got %d %#v\n", len(filterArgs), filterArgs)
	}

	if filterArgs[0] != filter.OrgID {
		t.Fatalf("expected OrgID %v got %v", filter.OrgID, filterArgs[0])
	}

	if filterArgs[1] != filter.GroupID {
		t.Fatalf("expected GroupID %v got %v", filter.OrgID, filterArgs[1])
	}

	if filterArgs[2] != filter.IgnoredValue {
		t.Fatalf("expected IgnoredValue %v got %v", filter.IgnoredValue, filterArgs[2])
	}

	if filterArgs[3] != filter.AfterScannedTime {
		t.Fatalf("expected AfterScannedTime %v got %v", filter.AfterScannedTime, filterArgs[3])
	}

	if filterArgs[4] != filter.BeforeScannedTime {
		t.Fatalf("expected BeforeScannedTime %v got %v", filter.BeforeScannedTime, filterArgs[4])
	}

	if filterArgs[5] != filter.AfterSeenTime {
		t.Fatalf("expected AfterSeenTime %v got %v", filter.AfterSeenTime, filterArgs[3])
	}

	if filterArgs[6] != filter.BeforeSeenTime {
		t.Fatalf("expected BeforeSeenTime %v got %v", filter.BeforeSeenTime, filterArgs[4])
	}

	if filterArgs[7] != filter.IsWildcardValue {
		t.Fatalf("expected IsWildcardValue %v got %v", filter.IsWildcardValue, filterArgs[5])
	}

	if filterArgs[8] != filter.IsHostedServiceValue {
		t.Fatalf("expected IsHostedServiceValue %v got %v", filter.IsHostedServiceValue, filterArgs[6])
	}

	if filterArgs[9] != convert.Reverse(filter.MatchesHost) {
		t.Fatalf("expected MatchesHost %v got %v", convert.Reverse(filter.MatchesHost), filterArgs[7])
	}

	if filterArgs[10] != convert.Reverse(filter.MatchesIP) {
		t.Fatalf("expected MatchesIP %v got %v", convert.Reverse(filter.MatchesIP), filterArgs[8])
	}

	if filterArgs[11] != filter.NSRecord {
		t.Fatalf("expected NSRecord %v got %v", filter.NSRecord, filterArgs[9])
	}

	if filterArgs[12] != filter.Start {
		t.Fatalf("expected Start %v got %v", filter.Start, filterArgs[10])
	}

	if filterArgs[13] != filter.Limit {
		t.Fatalf("expected Limit %v got %v", filter.Limit, filterArgs[11])
	}
}

func TestNew(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := address.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}
}

func TestGetHostList(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "hostlist"
	groupName := "hostlistgroup"

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
	_, count, err := service.Count(ctx, userContext, groupID)
	if err != nil {
		t.Fatalf("error getting empty count: %s\n", err)
	}

	if count != 0 {
		t.Fatalf("count should be zero for empty scangroup got: %d\n", count)
	}

	addresses := make(map[string]*am.ScanGroupAddress, 0)
	now := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		host := "www.example.com"
		ip := fmt.Sprintf("192.168.1.%d", i)
		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         host,
			IPAddress:           ip,
			AddressHash:         convert.HashAddress(ip, host),
			DiscoveryTime:       now,
			DiscoveredBy:        "input_list",
			LastScannedTime:     0,
			LastSeenTime:        0,
			ConfidenceScore:     0.0,
			UserConfidenceScore: 0.0,
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
		}

		addresses[a.AddressHash] = a
	}

	_, count, err = service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addreses: %s\n", err)
	}

	if count != 100 {
		t.Fatalf("error expected count to be 100, got: %d\n", count)
	}
	// test invalid (missing groupID) filter first
	filter := &am.ScanGroupAddressFilter{
		Start:   0,
		Limit:   10000,
		GroupID: groupID,
	}

	oid, hosts, err := service.GetHostList(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting host list: %v\n", err)
	}

	if oid != userContext.GetOrgID() {
		t.Fatalf("oid %v did not equal userContext id %v\n", oid, userContext.GetOrgID())
	}

	if len(hosts) != 1 {
		t.Fatalf("expected 1 host (multiple IPs) got %d\n", len(hosts))
	}

	if len(hosts[0].IPAddresses) != 100 {
		t.Fatalf("expected 100 IP addresses for host got %d\n", len(hosts[0].IPAddresses))
	}

	if len(hosts[0].AddressIDs) != 100 {
		t.Fatalf("expected 100 AddressIDs for host got %d\n", len(hosts[0].AddressIDs))
	}

}
func TestAdd(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

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
	_, count, err := service.Count(ctx, userContext, groupID)
	if err != nil {
		t.Fatalf("error getting empty count: %s\n", err)
	}

	if count != 0 {
		t.Fatalf("count should be zero for empty scangroup got: %d\n", count)
	}

	addresses := make(map[string]*am.ScanGroupAddress, 0)
	now := time.Now().UnixNano()
	for i := 0; i < 100; i++ {
		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         "",
			IPAddress:           fmt.Sprintf("192.168.1.%d", i),
			AddressHash:         convert.HashAddress(fmt.Sprintf("192.168.1.%d", i), ""),
			DiscoveryTime:       now,
			DiscoveredBy:        "input_list",
			LastScannedTime:     0,
			LastSeenTime:        0,
			ConfidenceScore:     0.0,
			UserConfidenceScore: 0.0,
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
		}

		addresses[a.AddressHash] = a
	}

	_, count, err = service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addreses: %s\n", err)
	}
	if count != 100 {
		t.Fatalf("error expected count to be 100, got: %d\n", count)
	}
	// test invalid (missing groupID) filter first
	filter := &am.ScanGroupAddressFilter{
		Start: 0,
		Limit: 10,
	}

	_, _, err = service.Get(ctx, userContext, filter)
	if err == nil {
		t.Fatalf("addresses did not return error with filter missing groupID")
	}

	filter.GroupID = groupID
	_, addrs, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %s\n", err)
	}

	if filter.Limit != len(addrs) {
		t.Fatalf("expected %d addresses with limit, got %d\n", filter.Limit, len(addrs))
	}

	filter.Limit = 100
	_, allAddresses, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting all addresses: %s\n", err)
	}

	// make list of all addressIDs for deletion
	addressIDs := make([]int64, len(allAddresses))
	for i := 0; i < len(allAddresses); i++ {
		addressIDs[i] = allAddresses[i].AddressID
	}

	if _, err := service.Delete(ctx, userContext, groupID, addressIDs); err != nil {
		t.Fatalf("error deleting all addresses: %s\n", err)
	}

	_, count, err = service.Count(ctx, userContext, groupID)
	if err != nil {
		t.Fatalf("error getting count: %s\n", err)
	}

	if count != 0 {
		t.Fatalf("error not all addresses were deleted got %d\n", count)
	}

}

func TestUpdate(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "updateaddress"
	groupName := "updateaddressgroup"

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

	now := time.Now().UnixNano()

	emptyAddress := &am.ScanGroupAddress{
		OrgID:           orgID,
		GroupID:         groupID,
		HostAddress:     "",
		IPAddress:       "",
		AddressHash:     convert.HashAddress("", ""),
		DiscoveryTime:   now,
		DiscoveredBy:    "input_list",
		LastSeenTime:    0,
		IsSOA:           false,
		IsWildcardZone:  false,
		IsHostedService: false,
		Ignored:         false,
	}
	emptyAddresses := make(map[string]*am.ScanGroupAddress, 1)

	emptyAddresses[emptyAddress.AddressHash] = emptyAddress
	if _, _, err := service.Update(ctx, userContext, emptyAddresses); err != address.ErrAddressMissing {
		t.Fatalf("did not get ErrAddressMissing when host/ip not set")
	}

	// test updating addresses
	updateAddress := &am.ScanGroupAddress{
		OrgID:           orgID,
		GroupID:         groupID,
		HostAddress:     "example.com",
		IPAddress:       "",
		AddressHash:     convert.HashAddress("", "example.com"),
		DiscoveryTime:   now,
		DiscoveredBy:    "input_list",
		LastSeenTime:    0,
		IsSOA:           false,
		IsWildcardZone:  false,
		IsHostedService: false,
		Ignored:         false,
		FoundFrom:       "",
		NSRecord:        1,
	}

	updateAddresses := make(map[string]*am.ScanGroupAddress, 1)
	updateAddresses[updateAddress.AddressHash] = updateAddress
	if _, _, err := service.Update(ctx, userContext, updateAddresses); err != nil {
		t.Fatalf("error creating address: %s\n", err)
	}

	filter := &am.ScanGroupAddressFilter{
		GroupID: groupID,
		Start:   0,
		Limit:   10,
	}

	_, returned, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting returned addresses")
	}
	compareAddresses(updateAddress, returned[0], t)

	// test updating last seen time
	now = time.Now().UnixNano()
	returned[0].LastSeenTime = now
	// various field updates:
	returned[0].ConfidenceScore = 99.9
	returned[0].UserConfidenceScore = 50.0
	returned[0].LastScannedTime = now
	returned[0].IsSOA = true
	returned[0].IsWildcardZone = true
	returned[0].IsHostedService = true

	returnMap := make(map[string]*am.ScanGroupAddress)
	returnMap[returned[0].AddressHash] = returned[0]
	if _, _, err := service.Update(ctx, userContext, returnMap); err != nil {
		t.Fatalf("error updating time for address: %s\n", err)
	}

	_, returned2, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting returned addresses after updating time")
	}
	compareAddresses(returned[0], returned2[0], t)

}

func TestIgnore(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "ignoreaddress"
	groupName := "ignoreaddressgroup"

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

	now := time.Now().UnixNano()

	addresses := make(map[string]*am.ScanGroupAddress, 0)
	for i := 0; i < 100; i++ {
		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         "",
			IPAddress:           fmt.Sprintf("192.168.1.%d", i),
			AddressHash:         convert.HashAddress(fmt.Sprintf("192.168.1.%d", i), ""),
			DiscoveryTime:       now,
			DiscoveredBy:        "input_list",
			LastScannedTime:     0,
			LastSeenTime:        0,
			ConfidenceScore:     0.0,
			UserConfidenceScore: 0.0,
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
		}

		addresses[a.AddressHash] = a
	}

	_, count, err := service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addreses: %s\n", err)
	}
	if count != 100 {
		t.Fatalf("error expected count to be 100, got: %d\n", count)
	}

	filter := &am.ScanGroupAddressFilter{
		GroupID: groupID,
		Start:   0,
		Limit:   100,
	}
	_, allAddresses, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting all addresses: %s\n", err)
	}

	// make list of all addressIDs for ignoring
	addressIDs := make([]int64, len(allAddresses))
	for i := 0; i < len(allAddresses); i++ {
		addressIDs[i] = allAddresses[i].AddressID
	}

	if _, err := service.Ignore(ctx, userContext, groupID, addressIDs, true); err != nil {
		t.Fatalf("error ignoring all addresses: %s\n", err)
	}

	filter.Limit = 100
	_, allAddresses, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting all addresses: %s\n", err)
	}

	for _, addr := range allAddresses {
		if addr.Ignored == false {
			t.Fatalf("error ignoring address: %v\n", addr.AddressID)
		}
	}

	if _, err := service.Ignore(ctx, userContext, groupID, addressIDs, false); err != nil {
		t.Fatalf("error ignoring all addresses: %s\n", err)
	}

	filter.Limit = 100
	_, allAddresses, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting all addresses: %s\n", err)
	}

	for _, addr := range allAddresses {
		if addr.Ignored == true {
			t.Fatalf("error ignoring address: %v\n", addr.AddressID)
		}
	}
}

func compareAddresses(e, r *am.ScanGroupAddress, t *testing.T) {
	if e.OrgID != r.OrgID {
		t.Fatalf("OrgID did not match expected: %v got: %v\n", e.OrgID, r.OrgID)
	}

	if e.GroupID != r.GroupID {
		t.Fatalf("GroupID did not match expected: %v got: %v\n", e.GroupID, r.GroupID)
	}

	if e.HostAddress != r.HostAddress {
		t.Fatalf("HostAddress did not match expected: %v got: %v\n", e.HostAddress, r.HostAddress)
	}

	if e.IPAddress != r.IPAddress {
		t.Fatalf("IPAddress did not match expected: %v got: %v\n", e.IPAddress, r.IPAddress)
	}

	if e.DiscoveryTime != r.DiscoveryTime {
		t.Fatalf("DiscoveryTime did not match expected: %v got: %v\n", e.DiscoveryTime, r.DiscoveryTime)
	}

	if e.DiscoveredBy != r.DiscoveredBy {
		t.Fatalf("DiscoveredBy did not match expected: %v got: %v\n", e.OrgID, r.OrgID)
	}

	if e.LastScannedTime != r.LastScannedTime {
		t.Fatalf("LastScannedTime did not match expected: %v got: %v\n", e.LastScannedTime, r.LastScannedTime)
	}

	if e.LastSeenTime != r.LastSeenTime {
		t.Fatalf("LastSeenTime did not match expected: %v got: %v\n", e.LastSeenTime, r.LastSeenTime)
	}

	if e.ConfidenceScore != r.ConfidenceScore {
		t.Fatalf("ConfidenceScore did not match expected: %v got: %v\n", e.ConfidenceScore, r.ConfidenceScore)
	}

	if e.UserConfidenceScore != r.UserConfidenceScore {
		t.Fatalf("UserConfidenceScore did not match expected: %v got: %v\n", e.UserConfidenceScore, r.UserConfidenceScore)
	}

	if e.IsSOA != r.IsSOA {
		t.Fatalf("IsSOA did not match expected: %v got: %v\n", e.IsSOA, r.IsSOA)
	}

	if e.IsWildcardZone != r.IsWildcardZone {
		t.Fatalf("IsWildcardZone did not match expected: %v got: %v\n", e.IsWildcardZone, r.IsWildcardZone)
	}

	if e.IsHostedService != r.IsHostedService {
		t.Fatalf("IsHostedService did not match expected: %v got: %v\n", e.IsHostedService, r.IsHostedService)
	}

	if e.Ignored != r.Ignored {
		t.Fatalf("Ignored did not match expected: %v got: %v\n", e.Ignored, r.Ignored)
	}
}
