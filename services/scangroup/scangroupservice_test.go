package scangroup_test

import (
	"context"
	"flag"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/secrets"

	"gopkg.linkai.io/v1/repos/am/amtest"

	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
)

var env string
var dbstring string

const serviceKey = "scangroupservice"

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
	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	orgName := "sgcreate"
	groupName := "sgcreategroup"

	auth := amtest.MockEmptyAuthorizer()

	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
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

	auth := amtest.MockEmptyAuthorizer()

	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
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

func TestPauseResume(t *testing.T) {
	ctx := context.Background()

	orgName := "sgpause"
	groupName := "sgpausegroup"

	auth := amtest.MockEmptyAuthorizer()

	service := scangroup.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer testForceCleanUp(db, orgID, orgName, t)

	ownerUserID := amtest.GetUserId(db, orgID, orgName, t)

	userContext := testUserContext(orgID, ownerUserID)
	group := testCreateNewGroup(orgID, ownerUserID, groupName)

	_, gid, err := service.Create(ctx, userContext, group)
	if err != nil {
		t.Fatalf("error creating group: %s\n", err)
	}

	_, gid, err = service.Pause(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error pausing group: %s\n", err)
	}

	_, paused, err := service.Get(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting paused group: %s\n", err)
	}

	if paused.Paused == false {
		t.Fatalf("scan group was not paused: %v\n", paused.Paused)
	}

	_, gid, err = service.Resume(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error resuming group: %s\n", err)
	}

	_, resumed, err := service.Get(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting paused group: %s\n", err)
	}

	if resumed.Paused == true {
		t.Fatalf("scan group was not resumed: %v\n", resumed.Paused)
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
