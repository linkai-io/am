package scangroup_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/amtest"

	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
)

var dbstring = os.Getenv("SCANGROUPSERVICE_DB_STRING")

func TestNew(t *testing.T) {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	orgName := "sgcreate"
	groupName := "sgcreategroup"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer testForceCleanUp(db, orgID, orgName, t)

	ownerUserID := amtest.GetUserId(db, orgID, orgName, t)

	userContext := testUserContext(orgID, ownerUserID)
	group := testCreateNewGroup(orgID, ownerUserID, groupName)

	oid, gid, err := service.Create(ctx, userContext, group)
	if err != nil {
		t.Fatalf("error creating group: %s\n", err)
	}

	if orgID != oid {
		t.Fatalf("error orgID did not match expected: %d got: %d\n", orgID, oid)
	}

	if gid == 0 {
		t.Fatalf("groupid returned was 0, gid:%d\n", gid)
	}

	// test create same group/version returns error
	if _, _, err = service.Create(ctx, userContext, group); err == nil {
		t.Fatalf("did not get error recreating same group")
	}

	doid, dgid, err := service.Delete(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error deleting scan group: %s\n", err)
	}

	if orgID != doid {
		t.Fatalf("deleted org id did not match orgid: expected: %d got: %d\n", orgID, doid)
	}

	if gid != dgid {
		t.Fatalf("deleted group id did not match: expected: %d got: %d\n", gid, dgid)
	}

	if _, gid, err = service.Create(ctx, userContext, group); err != nil {
		_, groups, _ := service.Groups(ctx, userContext)
		for g := range groups {
			t.Logf("%#v\n", g)
		}
		t.Fatalf("error re-creating same group after delete: %s\n", err)
	}

	_, returned, err := service.Get(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting group: %s\n", err)
	}

	_, returnedName, err := service.GetByName(ctx, userContext, returned.GroupName)
	if err != nil {
		t.Fatalf("error getting group by name: %s\n", err)
	}

	testCompareGroups(returned, returnedName, t)

	// test update
	t.Logf("%#v\n", returned)
	returned.ModifiedTime = time.Now().UnixNano()
	returned.GroupName = "modified group"
	returned.ModuleConfigurations = &am.ModuleConfiguration{NSModule: &am.NSModuleConfig{Name: "NS"}}
	t.Logf("%#v\n", returned)

	_, ugid, err := service.Update(ctx, userContext, returned)
	if err != nil {
		t.Fatalf("error updating returned group: %s\n", err)
	}

	if gid != ugid {
		t.Fatalf("error groupid changed expected %d got %d\n", gid, ugid)
	}

	// test we can't access by old name
	if _, _, err = service.GetByName(ctx, userContext, returnedName.GroupName); err == nil {
		t.Fatalf("should have got error getting group by old name: %s\n", returnedName.GroupName)
	}

	_, mod, err := service.GetByName(ctx, userContext, "modified group")
	if err != nil {
		t.Fatalf("error getting modified group by name: %s\n", err)
	}

	if returnedName.ModifiedTime == mod.ModifiedTime {
		t.Fatalf("modified time was not updated, expected %d got %d\n", mod.ModifiedTime, returned.ModifiedTime)
	}

	if mod.ModuleConfigurations == nil || mod.ModuleConfigurations.NSModule == nil || mod.ModuleConfigurations.NSModule.Name != "NS" {
		t.Fatalf("module configurations was not returned properly: %#v\n", mod.ModuleConfigurations)
	}
}

func TestGetGroups(t *testing.T) {
	ctx := context.Background()

	orgName := "sggetgroups"
	groupName := "sggetgroupsgroup"
	count := 10

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer testForceCleanUp(db, orgID, orgName, t)

	ownerUserID := amtest.GetUserId(db, orgID, orgName, t)
	userContext := testUserContext(orgID, ownerUserID)

	GIDs := make([]int, count)

	for i := 0; i < count; i++ {
		groupName = fmt.Sprintf("%s%d", groupName, i)

		group := testCreateNewGroup(orgID, ownerUserID, groupName)
		oid, gid, err := service.Create(ctx, userContext, group)
		if err != nil {
			t.Fatalf("error creating new group for %d: %s\n", i, err)
		}
		if orgID != oid {
			t.Fatalf("error mismatched orgID: expected %d got %d\n", orgID, oid)
		}

		GIDs[i] = int(gid)
	}

	_, groups, err := service.Groups(ctx, userContext)
	if err != nil {
		t.Fatalf("error getting groups: %s\n", err)
	}

	groupGIDs := make([]int, count)
	for i, group := range groups {
		groupGIDs[i] = int(group.GroupID)
	}
	amtest.SortEqualInt(GIDs, groupGIDs, t)

	for i := 0; i < count; i++ {
		if _, _, err := service.Get(ctx, userContext, int(GIDs[i])); err != nil {
			t.Fatalf("error getting group %d: %s\n", GIDs[i], err)
		}
	}

	// test getting invalid group id returns error
	if _, _, err := service.Get(ctx, userContext, -1); err == nil {
		t.Fatalf("expected error getting invalid group id\n")
	} else {
		if err != am.ErrScanGroupNotExists {
			t.Fatalf("expected errscangroupnotexists got: %s\n", err)
		}
	}
}

func TestAddAddresses(t *testing.T) {
	ctx := context.Background()

	orgName := "sgaddaddress"
	groupName := "sgaddaddressgroup"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer testForceCleanUp(db, orgID, orgName, t)

	ownerUserID := amtest.GetUserId(db, orgID, orgName, t)
	userContext := testUserContext(orgID, ownerUserID)
	group := testCreateNewGroup(orgID, ownerUserID, groupName)

	_, gid, err := service.Create(ctx, userContext, group)
	if err != nil {
		t.Fatalf("error creating group: %s\n", err)
	}

	header := &am.ScanGroupAddressHeader{Ignored: false, GroupID: gid, AddedBy: "ns_queries"}
	count := 200
	ips := make([]string, count)
	for i := 0; i < count; i++ {
		ips[i] = fmt.Sprintf("192.168.1.%d", i)
	}

	if _, err := service.AddAddresses(ctx, userContext, header, ips); err != nil {
		t.Fatalf("error adding addresses: %s\n", err)
	}

	orgName2 := "sgaddaddress2"

	amtest.CreateOrg(db, orgName2, t)
	orgID2 := amtest.GetOrgID(db, orgName2, t)
	defer testForceCleanUp(db, orgID2, orgName2, t)

	ownerUserID2 := amtest.GetUserId(db, orgID2, orgName2, t)
	userContext2 := testUserContext(orgID2, ownerUserID2)
	group2 := testCreateNewGroup(orgID2, ownerUserID2, groupName)

	_, gid2, err := service.Create(ctx, userContext2, group2)
	if err != nil {
		t.Fatalf("error creating 2nd group: %s\n", err)
	}

	header2 := &am.ScanGroupAddressHeader{Ignored: false, GroupID: gid2, AddedBy: "ns_queries"}
	count2 := 200
	ips2 := make([]string, count2)
	for i := 0; i < count2; i++ {
		ips2[i] = fmt.Sprintf("192.168.1.%d", i)
	}

	if _, err := service.AddAddresses(ctx, userContext2, header2, ips2); err != nil {
		t.Fatalf("error adding addresses: %s\n", err)
	}

	// Test address count/ Addresses
	_, returnedCount, err := service.AddressCount(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting address count: %s\n", err)
	}
	if count != returnedCount {
		t.Fatalf("expected %d got %d addresses\n", count, returnedCount)
	}

	filter := &am.ScanGroupAddressFilter{GroupID: gid, Start: 0, Limit: 100, WithDeleted: true, DeletedValue: false, WithIgnored: false}

	_, addresses, err := service.Addresses(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses for first group: %s\n", err)
	}

	if len(addresses) != 100 {
		t.Fatalf("expected 100 addresses got: %d\n", len(addresses))
	}

	_, returnedCount, err = service.AddressCount(ctx, userContext2, gid2)
	if err != nil {
		t.Fatalf("error getting address count: %s\n", err)
	}
	if count != returnedCount {
		t.Fatalf("expected %d got %d addresses\n", count, returnedCount)
	}

	filter2 := &am.ScanGroupAddressFilter{GroupID: gid2, Start: 0, Limit: 100, WithDeleted: true, DeletedValue: false, WithIgnored: false}
	_, addresses, err = service.Addresses(ctx, userContext2, filter2)
	if err != nil {
		t.Fatalf("error getting addresses for second group: %s\n", err)
	}
	if len(addresses) != 100 {
		t.Fatalf("expected 100 addresses got: %d\n", len(addresses))
	}
}

func TestUpdateAddresses(t *testing.T) {
	ctx := context.Background()

	orgName := "sgupdateaddress"
	groupName := "sgupdateaddressgroup"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer testForceCleanUp(db, orgID, orgName, t)

	ownerUserID := amtest.GetUserId(db, orgID, orgName, t)
	userContext := testUserContext(orgID, ownerUserID)
	group := testCreateNewGroup(orgID, ownerUserID, groupName)

	_, gid, err := service.Create(ctx, userContext, group)
	if err != nil {
		t.Fatalf("error creating group: %s\n", err)
	}

	header := &am.ScanGroupAddressHeader{Ignored: false, GroupID: gid, AddedBy: "ns_queries"}
	count := 20
	ips := make([]string, count)
	for i := 0; i < count; i++ {
		ips[i] = fmt.Sprintf("192.168.1.%d", i)
	}

	if _, err := service.AddAddresses(ctx, userContext, header, ips); err != nil {
		t.Fatalf("error adding addresses: %s\n", err)
	}

	filter := &am.ScanGroupAddressFilter{GroupID: gid, Start: 0, Limit: 10, WithDeleted: true, DeletedValue: false}
	_, addresses, err := service.Addresses(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %s\n", err)
	}

	// test ignoring
	addressIDs := make(map[int64]bool, 10)
	for i := 0; i < 10; i++ {
		id := addresses[i].AddressID
		addressIDs[id] = true
	}

	if _, err := service.IgnoreAddresses(ctx, userContext, gid, addressIDs); err != nil {
		t.Fatalf("error ignoring addresses: %s\n", err)
	}

	filter.WithDeleted = false
	filter.WithIgnored = true
	filter.IgnoredValue = true
	_, addresses, err = service.Addresses(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %s\n", err)
	}

	if len(addresses) != 10 {
		t.Fatalf("expected 10 ignored addresses got: %d\n", len(addresses))
	}

	// delete ignored addresses
	if _, err := service.DeleteAddresses(ctx, userContext, gid, addressIDs); err != nil {
		t.Fatalf("error deleting ignored addresses: %s\n", err)
	}

	filter.WithDeleted = true
	filter.DeletedValue = true
	_, addresses, err = service.Addresses(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %s\n", err)
	}

	if len(addresses) != 10 {
		t.Fatalf("expected 10 deleted addresses got: %d\n", len(addresses))
	}

	filter.WithDeleted = false
	filter.WithIgnored = false
	filter.Limit = 20
	_, addresses, err = service.Addresses(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error getting addresses: %s\n", err)
	}

	if len(addresses) != 20 {
		t.Fatalf("expected 20 deleted addresses got: %d\n", len(addresses))
	}
}

func testCompareGroups(group1, group2 *am.ScanGroup, t *testing.T) {
	if group1.CreatedBy != group2.CreatedBy {
		t.Fatalf("created by was different, %d and %d\n", group1.CreatedBy, group2.CreatedBy)
	}

	if group1.ModifiedBy != group2.ModifiedBy {
		t.Fatalf("modified by was different, %d and %d\n", group1.ModifiedBy, group2.ModifiedBy)
	}

	if group1.CreationTime != group2.CreationTime {
		t.Fatalf("creation time by was different, %d and %d\n", group1.CreationTime, group2.CreationTime)
	}

	if group1.ModifiedTime != group2.ModifiedTime {
		t.Fatalf("ModifiedTime by was different, %d and %d\n", group1.CreationTime, group2.CreationTime)
	}

	if group1.GroupID != group2.GroupID {
		t.Fatalf("GroupID by was different, %d and %d\n", group1.GroupID, group2.GroupID)
	}

	if group1.OrgID != group2.OrgID {
		t.Fatalf("OrgID by was different, %d and %d\n", group1.OrgID, group2.OrgID)
	}

	if group1.GroupName != group2.GroupName {
		t.Fatalf("GroupName by was different, %s and %s\n", group1.GroupName, group2.GroupName)
	}

	if string(group1.OriginalInput) != string(group2.OriginalInput) {
		t.Fatalf("OriginalInput by was different, %s and %s\n", string(group1.OriginalInput), string(group2.OriginalInput))
	}
}

func testForceCleanUp(db *pgx.ConnPool, orgID int, orgName string, t *testing.T) {
	db.Exec("delete from am.scan_group_addresses where organization_id=$1", orgID)
	db.Exec("delete from am.scan_group where organization_id=$1", orgID)
	amtest.DeleteOrg(db, orgName, t)
}

func testCreateNewGroup(orgID, userID int, groupName string) *am.ScanGroup {
	now := time.Now().UnixNano()
	group := &am.ScanGroup{
		OrgID:                orgID,
		GroupName:            groupName,
		CreationTime:         now,
		CreatedBy:            userID,
		ModifiedTime:         now,
		ModifiedBy:           userID,
		ModuleConfigurations: &am.ModuleConfiguration{},
		OriginalInput:        []byte("192.168.0.1"),
	}

	return group
}

func testUserContext(orgID, userID int) *mock.UserContext {
	userContext := &mock.UserContext{}
	userContext.GetOrgIDFn = func() int {
		return orgID
	}

	userContext.GetUserIDFn = func() int {
		return userID
	}

	return userContext
}
