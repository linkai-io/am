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
	_, count, err := service.Count(ctx, userContext, groupID)
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
			LastScannedTime: 0,
			LastSeenTime:    0,
			ConfidenceScore: 0.0,
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
		DiscoveryTime:   now,
		DiscoveredBy:    "input_list",
		LastSeenTime:    0,
		IsSOA:           false,
		IsWildcardZone:  false,
		IsHostedService: false,
		Ignored:         false,
	}
	emptyAddresses := make([]*am.ScanGroupAddress, 1)
	emptyAddresses[0] = emptyAddress
	if _, _, err := service.Update(ctx, userContext, emptyAddresses); err != address.ErrAddressMissing {
		t.Fatalf("did not get ErrAddressMissing when host/ip not set")
	}

	// test updating addresses
	updateAddress := &am.ScanGroupAddress{
		OrgID:           orgID,
		GroupID:         groupID,
		HostAddress:     "example.com",
		IPAddress:       "",
		DiscoveryTime:   now,
		DiscoveredBy:    "input_list",
		LastSeenTime:    0,
		IsSOA:           false,
		IsWildcardZone:  false,
		IsHostedService: false,
		Ignored:         false,
	}

	updateAddresses := make([]*am.ScanGroupAddress, 1)
	updateAddresses[0] = updateAddress
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
	returned[0].LastScannedTime = now
	returned[0].IsSOA = true
	returned[0].IsWildcardZone = true
	returned[0].IsHostedService = true

	if _, _, err := service.Update(ctx, userContext, returned); err != nil {
		t.Fatalf("error updating time for address: %s\n", err)
	}

	_, returned2, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting returned addresses after updating time")
	}
	compareAddresses(returned[0], returned2[0], t)
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
