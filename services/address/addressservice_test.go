package address_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/miekg/dns"

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

	orgName := "hostlistt"
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
	now := time.Now()
	for i := 0; i < 100; i++ {
		host := "www.example.com"
		ip := fmt.Sprintf("192.168.1.%d", i)

		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         host,
			IPAddress:           ip,
			AddressHash:         convert.HashAddress(ip, host),
			DiscoveryTime:       now.Add(time.Hour * time.Duration(-i*2)).UnixNano(),
			DiscoveredBy:        "input_list",
			LastScannedTime:     now.Add(time.Hour * time.Duration(-i)).UnixNano(),
			LastSeenTime:        now.Add(time.Hour * time.Duration(-i)).UnixNano(),
			ConfidenceScore:     100,
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
		Start:   0,
		Limit:   10,
		Filters: &am.FilterType{},
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
		Filters: &am.FilterType{},
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

func TestCreateUpdateSmall(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "updateaddresssmall"
	groupName := "updateaddresssmallgroup"

	auth := amtest.MockEmptyAuthorizer()

	service := address.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateSmallOrg(db, orgName, t)
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
		ip := fmt.Sprintf("192.168.1.%d", i)
		host := fmt.Sprintf("%d.example.com", i)
		hash := convert.HashAddress(ip, host)
		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         host,
			IPAddress:           ip,
			AddressHash:         hash,
			DiscoveryTime:       now,
			DiscoveredBy:        "input_list",
			LastScannedTime:     0,
			LastSeenTime:        0,
			ConfidenceScore:     100,
			UserConfidenceScore: 0.0,
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
		}

		if i >= 80 && i <= 84 {
			a.NSRecord = int32(dns.TypeMX)
			a.ConfidenceScore = 0
		} else if i >= 85 && i < 90 {
			a.NSRecord = int32(dns.TypeNS)
			a.ConfidenceScore = 0
		}
		addresses[a.AddressHash] = a
	}

	_, count, err = service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addresses: %s\n", err)
	}

	if count != 100 {
		t.Fatalf("error expected count to be 100, got: %d\n", count)
	}

	filter := &am.ScanGroupAddressFilter{
		OrgID:   orgID,
		Start:   0,
		Limit:   100,
		GroupID: groupID,
		Filters: &am.FilterType{},
	}

	_, addrs, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("addresses returned error: %v\n", err)
	}

	if len(addrs) != 35 {
		t.Fatalf("expected 35 addresses, got: %d\n", len(addrs))
	}

	mxCount := 0
	nsCount := 0
	for _, addr := range addrs {
		if addr.NSRecord == int32(dns.TypeMX) {
			mxCount++
		} else if addr.NSRecord == int32(dns.TypeNS) {
			nsCount++
		}
	}

	if mxCount != 5 || nsCount != 5 {
		t.Fatalf("expected 5, 5 for mx/ns count, got %v %v\n", mxCount, nsCount)
	}

	// test adding new addresses (ips) but same hostnames
	addresses = make(map[string]*am.ScanGroupAddress, 0)
	now = time.Now().UnixNano()
	for i := 0; i < 10; i++ {
		ip := fmt.Sprintf("192.168.2.%d", i) // different ip
		a := addrs[i]
		a.DiscoveryTime = now
		a.IPAddress = ip
		a.AddressHash = convert.HashAddress(ip, addrs[i].HostAddress)
		a.DiscoveredBy = "ns_query_name_to_ip"

		addresses[a.AddressHash] = a
	}

	_, count, err = service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addresses: %s\n", err)
	}

	if count != 10 {
		t.Fatalf("error expected count to be 10, got: %d\n", count)
	}

	_, addrs, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("addresses returned error: %v\n", err)
	}

	if len(addrs) != 45 {
		t.Fatalf("expected 45 addresses, got: %d\n", len(addrs))
	}
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
		Filters: &am.FilterType{},
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

func TestGetAddressFilters(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "addressfilters"
	groupName := "addressfilters"

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

	addresses := make(map[string]*am.ScanGroupAddress, 0)
	now := time.Now()
	for i := 0; i < 100; i++ {
		host := fmt.Sprintf("%d.example.com", i)
		ip := fmt.Sprintf("192.168.1.%d", i)

		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         host,
			IPAddress:           ip,
			AddressHash:         convert.HashAddress(ip, host),
			DiscoveryTime:       now.Add(time.Hour * time.Duration(-i*2)).UnixNano(),
			DiscoveredBy:        "input_list",
			LastScannedTime:     now.Add(time.Hour * time.Duration(-i)).UnixNano(),
			LastSeenTime:        now.Add(time.Hour * time.Duration(-i)).UnixNano(),
			ConfidenceScore:     float32(i),
			UserConfidenceScore: float32(i),
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
			NSRecord:            int32(i),
		}

		addresses[a.AddressHash] = a
	}

	_, _, err := service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addresses: %v\n", err)
	}

	filter := &am.ScanGroupAddressFilter{
		OrgID:   orgID,
		GroupID: groupID,
		Start:   0,
		Limit:   1000,
		Filters: &am.FilterType{},
	}

	_, returned, err := service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != len(addresses) {
		t.Fatalf("expected %d returned got %d\n", len(addresses), len(returned))
	}

	filter.Filters.AddBool("wildcard", true)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 0 {
		t.Fatalf("expected 0 returned got: %d\n", len(returned))
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddBool("hosted", true)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 0 {
		t.Fatalf("expected 0 returned got: %d\n", len(returned))
	}

	// test scanned time
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("after_scanned_time", now.Add(time.Hour*time.Duration(-5)).UnixNano())
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("before_scanned_time", now.Add(time.Hour*time.Duration(-4)).UnixNano())
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 95 {
		t.Fatalf("expected 95 returned got: %d\n", len(returned))
	}

	// test seen time
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("after_seen_time", now.Add(time.Hour*time.Duration(-5)).UnixNano())
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("before_seen_time", now.Add(time.Hour*time.Duration(-4)).UnixNano())
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 95 {
		t.Fatalf("expected 95 returned got: %d\n", len(returned))
	}
	// test discovered time
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("after_discovered_time", now.Add(time.Hour*time.Duration(-5*2)).UnixNano())
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt64("before_discovered_time", now.Add(time.Hour*time.Duration(-4*2)).UnixNano())
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 95 {
		t.Fatalf("expected 95 returned got: %d\n", len(returned))
	}

	// test confidence
	filter.Filters = &am.FilterType{}
	filter.Filters.AddFloat32("above_confidence", 94)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}
	filter.Filters = &am.FilterType{}
	filter.Filters.AddFloat32("below_confidence", 5)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}
	// test user confidence
	filter.Filters = &am.FilterType{}
	filter.Filters.AddFloat32("above_user_confidence", 94)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}
	filter.Filters = &am.FilterType{}
	filter.Filters.AddFloat32("below_user_confidence", 5)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 5 {
		t.Fatalf("expected 5 returned got: %d\n", len(returned))
	}

	// test ns record
	filter.Filters = &am.FilterType{}
	filter.Filters.AddInt32("ns_record", 1)
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 1 {
		t.Fatalf("expected 1 returned got: %d\n", len(returned))
	}
	// test host address
	filter.Filters = &am.FilterType{}
	filter.Filters.AddString("host_address", "1.example.com")
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 1 {
		t.Fatalf("expected 1 returned got: %d\n", len(returned))
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddString("starts_host_address", "1")
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 11 {
		t.Fatalf("expected 11 returned got: %d\n", len(returned))
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddString("ends_host_address", "example.com")
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 100 {
		t.Fatalf("expected 100 returned got: %d\n", len(returned))
	}

	filter.Filters = &am.FilterType{}
	filter.Filters.AddString("ip_address", "192.168.1.1")
	_, returned, err = service.Get(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %v\n", err)
	}

	if len(returned) != 1 {
		t.Fatalf("expected 1 returned got: %d\n", len(returned))
	}
}

func TestOrgStats(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "addressorgstats"
	groupName := "addressorgstats"

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

	addresses := make(map[string]*am.ScanGroupAddress, 0)
	now := time.Now()
	for i := 0; i < 100; i++ {
		discovered := am.DiscoveryNSInputList
		host := fmt.Sprintf("%d.example.com", i)
		ip := fmt.Sprintf("192.168.1.%d", i)
		if i%10 == 0 {
			discovered = am.DiscoveryBruteSubDomain
		}
		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         host,
			IPAddress:           ip,
			AddressHash:         convert.HashAddress(ip, host),
			DiscoveryTime:       now.Add(time.Hour * time.Duration(-i*2)).UnixNano(),
			DiscoveredBy:        discovered,
			LastScannedTime:     now.Add(time.Hour * time.Duration(-i)).UnixNano(),
			LastSeenTime:        now.Add(time.Hour * time.Duration(-i)).UnixNano(),
			ConfidenceScore:     100,
			UserConfidenceScore: 100,
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
			NSRecord:            int32(i),
		}

		addresses[a.AddressHash] = a
	}

	_, _, err := service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addresses: %v\n", err)
	}

	// need to run aggregate functions first
	amtest.RunAggregates(db, t)
	oid, orgStats, err := service.OrgStats(ctx, userContext)
	if err != nil {
		t.Fatalf("error getting stats %v", err)
	}

	if len(orgStats) != 1 {
		t.Fatalf("error getting one scan group worth of stats, got %d\n", len(orgStats))
	}

	if oid != userContext.GetOrgID() {
		t.Fatalf("error org id mismatch")
	}

	var brute int32
	var input int32
	for i, s := range orgStats[0].DiscoveredBy {
		if s == am.DiscoveryBruteSubDomain {
			brute = orgStats[0].DiscoveredByCount[i]
		}
		if s == am.DiscoveryNSInputList {
			input = orgStats[0].DiscoveredByCount[i]
		}
	}
	if brute != 10 || input != 90 {
		t.Fatalf("expected 10 brute findings, and 90 input list got: %d %d", brute, input)
	}

	if orgStats[0].ConfidentTotal != 100 {
		t.Fatalf("expected 100 confident hosts got %d\n", orgStats[0].ConfidentTotal)
	}

	for k, v := range orgStats[0].Aggregates {
		t.Logf("%#v %#v\n", k, v)
		switch k {
		case "discovery_day":
		case "scanned_day":
		case "seen_day":
		case "discovery_trihourly":
		case "scanned_trihourly":
		case "seen_trihourly":
			if len(v.Time) == 0 {
				t.Fatalf("%s expected more than 0 results", k)
			}
			if len(v.Time) != len(v.Count) {
				t.Fatalf("%s expected count and time to match got %d != %d\n", k, len(v.Time), len(v.Count))
			}
		}
	}

	if orgStats[0].GroupID != groupID && orgStats[0].OrgID != orgID {
		t.Fatalf("org/group not set")
	}

	_, gstats, err := service.GroupStats(ctx, userContext, orgStats[0].GroupID)
	if err != nil {
		t.Fatalf("error getting group stats: %v\n", err)
	}

	if gstats.ConfidentTotal != orgStats[0].ConfidentTotal {
		t.Fatalf("same group returned different confident total")
	}

	if len(orgStats[0].Aggregates) != len(gstats.Aggregates) {
		t.Fatalf("same group should return same len of aggregates %d %d", len(orgStats[0].Aggregates), len(gstats.Aggregates))
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

	if e.DiscoveryTime/1000 != r.DiscoveryTime/1000 {
		t.Fatalf("DiscoveryTime did not match expected: %v got: %v\n", e.DiscoveryTime, r.DiscoveryTime)
	}

	if e.DiscoveredBy != r.DiscoveredBy {
		t.Fatalf("DiscoveredBy did not match expected: %v got: %v\n", e.OrgID, r.OrgID)
	}

	if e.LastScannedTime/1000 != r.LastScannedTime/1000 {
		t.Fatalf("LastScannedTime did not match expected: %v got: %v\n", e.LastScannedTime, r.LastScannedTime)
	}

	if e.LastSeenTime/1000 != r.LastSeenTime/1000 {
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

func TestPopulateData(t *testing.T) {
	t.Skip("uncomment to populate db")
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "testpopulatedata7"
	groupName := "testpopulatedata7"

	auth := amtest.MockEmptyAuthorizer()

	service := address.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	//defer amtest.DeleteOrg(db, orgName, t)

	groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
	userContext := amtest.CreateUserContext(orgID, 1)

	addresses := make(map[string]*am.ScanGroupAddress, 0)
	now := time.Now()
	for i := 0; i < 100000; i++ {
		host := fmt.Sprintf("%d.example.com", i)
		ip := fmt.Sprintf("192.168.1.%d", i)

		a := &am.ScanGroupAddress{
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         host,
			IPAddress:           ip,
			AddressHash:         convert.HashAddress(ip, host),
			DiscoveryTime:       now.Add(time.Minute * time.Duration(-i*2)).UnixNano(),
			DiscoveredBy:        "input_list",
			LastScannedTime:     now.Add(time.Minute * time.Duration(-i)).UnixNano(),
			LastSeenTime:        now.Add(time.Minute * time.Duration(-i)).UnixNano(),
			ConfidenceScore:     100,
			UserConfidenceScore: 0,
			IsSOA:               false,
			IsWildcardZone:      false,
			IsHostedService:     false,
			Ignored:             false,
			NSRecord:            1,
		}

		addresses[a.AddressHash] = a
	}

	_, _, err := service.Update(ctx, userContext, addresses)
	if err != nil {
		t.Fatalf("error adding addresses: %v\n", err)
	}
}
