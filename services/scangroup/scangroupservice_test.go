package scangroup_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/amtest"

	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
)

func TestNew(t *testing.T) {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing scangroup service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	orgName := "sgcreate"
	groupName := "sgcreategroup"
	versionName := "sgcreateversion"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int32, resource, action string) error {
		return nil
	}
	service := scangroup.New(auth)

	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
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
	group, version := testCreateNewGroup(orgID, ownerUserID, groupName, versionName)

	oid, gid, gvid, err := service.Create(ctx, userContext, group, version)
	if err != nil {
		t.Fatalf("error creating group: %s\n", err)
	}

	if orgID != oid {
		t.Fatalf("error orgID did not match expected: %d got: %d\n", orgID, oid)
	}

	if gid == 0 || gvid == 0 {
		t.Fatalf("groupid/gvid returned was 0, gid:%d gvid:%d\n", gid, gvid)
	}

	goid, gversion, err := service.GetVersionByName(ctx, userContext, gid, versionName)
	if err != nil {
		t.Fatalf("error getting version by name: %s\n", err)
	}

	if orgID != goid {
		t.Fatalf("orgid does not match get version orgID: expected: %d got %d\n", orgID, goid)
	}

	if versionName != gversion.VersionName {
		t.Fatalf("version name does not match: expected %s got %s\n", versionName, gversion.VersionName)
	}

	if gvid != gversion.GroupVersionID {
		t.Fatalf("group version id returned incorrect expected: %d got: %d\n", gvid, gversion.GroupVersionID)
	}

	t.Logf("ORGID: %d GROUPID: %d GROUPVERSIONID: %d\n", orgID, gid, gvid)

	doid, dgid, err := service.Delete(ctx, userContext, gid)
	if err == nil {
		t.Fatalf("did not get error deleting a scangroup that was referenced from a version\n")
	}

	t.Logf("version id: %d\n", gversion.GroupVersionID)

	if _, _, _, err = service.DeleteVersion(ctx, userContext, gversion.GroupID, gversion.GroupVersionID, gversion.VersionName); err != nil {
		t.Fatalf("error deleting version: %s\n", err)
	}

	doid, dgid, err = service.Delete(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error deleting scan group: %s\n", err)
	}

	if orgID != doid {
		t.Fatalf("deleted org id did not match orgid: expected: %d got: %d\n", orgID, doid)
	}

	if gid != dgid {
		t.Fatalf("deleted group id did not match: expected: %d got: %d\n", gid, dgid)
	}
}

func testForceCleanUp(db *pgx.ConnPool, orgID int32, orgName string, t *testing.T) {
	db.Exec("delete from am.scan_group_versions where organization_id=$1", orgID)
	db.Exec("delete from am.scan_group where organization_id=$1", orgID)
	amtest.DeleteOrg(db, orgName, t)
}

func testCreateNewGroup(orgID, userID int32, groupName, versionName string) (*am.ScanGroup, *am.ScanGroupVersion) {
	group := &am.ScanGroup{
		OrgID:         orgID,
		GroupName:     groupName,
		CreationTime:  time.Now().UnixNano(),
		CreatedBy:     userID,
		OriginalInput: []byte("192.168.0.1"),
	}
	version := &am.ScanGroupVersion{
		OrgID:                orgID,
		VersionName:          versionName,
		CreationTime:         time.Now().UnixNano(),
		CreatedBy:            userID,
		ModuleConfigurations: &am.ModuleConfiguration{},
	}
	return group, version
}

func testUserContext(orgID, userID int32) *mock.UserContext {
	userContext := &mock.UserContext{}
	userContext.GetOrgIDFn = func() int32 {
		return orgID
	}

	userContext.GetUserIDFn = func() int32 {
		return userID
	}

	return userContext
}
