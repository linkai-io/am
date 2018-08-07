package organization_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"gopkg.linkai.io/v1/repos/am/amtest"

	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/services/organization"
)

var dbstring = os.Getenv("ORGSERVICE_DB_STRING")

func TestNew(t *testing.T) {
	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()
	service := organization.New(roleManager, auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	orgName := "orgorgcreate"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := &am.Organization{}

	_, _, _, _, err := service.Create(ctx, userContext, org)
	if err == nil {
		t.Fatalf("did not get error creating invalid organization\n")
	}

	org = amtest.CreateOrgInstance(orgName)
	defer amtest.DeleteOrg(db, orgName, t)

	_, _, ocid, ucid, err := service.Create(ctx, userContext, org)
	if err != nil {
		t.Fatalf("error creating organization: %s\n", err)
	}

	if ocid == "" || ucid == "" {
		t.Fatalf("ocid: %s ucid: %s was empty\n", ocid, ucid)
	}

	_, returned, err := service.Get(ctx, userContext, orgName)
	if err != nil {
		t.Fatalf("error getting organization by name: %s\n", err)
	}
	amtest.TestCompareOrganizations(org, returned, t)

	var oid int
	oid, returned, err = service.GetByCID(ctx, userContext, ocid)
	if err != nil {
		t.Fatalf("error getting organization by cid: %s\n", err)
	}
	amtest.TestCompareOrganizations(org, returned, t)

	_, returned, err = service.GetByID(ctx, userContext, oid)
	if err != nil {
		t.Fatalf("error getting organization by cid: %s\n", err)
	}
	amtest.TestCompareOrganizations(org, returned, t)

	filter := &am.OrgFilter{
		Start: 0,
		Limit: 10,
	}
	orgs, err := service.List(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error listing organizations: %s\n", err)
	}

	if len(orgs) != 3 {
		t.Fatalf("expected 3 organization in list, got: %d\n", len(orgs))
	}
	amtest.TestCompareOrganizations(org, orgs[2], t)
}

func TestCreateRoleFail(t *testing.T) {
	ctx := context.Background()

	orgName := "orgcreaterolefail"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()
	roleManager.CreateRoleFn = func(role *am.Role) (string, error) {
		return "", errors.New("unable to add role")
	}

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := amtest.CreateOrgInstance(orgName)

	if _, _, _, _, err := service.Create(ctx, userContext, org); err == nil {
		t.Fatalf("role manager did not throw error\n")
	}
	_, _, err := service.Get(ctx, userContext, orgName)
	if err == nil {
		t.Fatalf("error role manager failure did not cause org to be deleted")
	}

	defer amtest.DeleteOrg(db, orgName, t)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	orgName := "orgorgdelete"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := amtest.CreateOrgInstance(orgName)

	if _, _, _, _, err := service.Create(ctx, userContext, org); err != nil {
		t.Fatalf("error creating organization: %s\n", err)
	}
	defer amtest.DeleteOrg(db, orgName, t)

	oid, _, err := service.Get(ctx, userContext, orgName)
	if err != nil {
		t.Fatalf("error getting organization: %s\n", err)
	}

	if _, err := service.Delete(ctx, userContext, oid); err != nil {
		t.Fatalf("error deleting organization: %s\n", err)
	}

}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	orgName := "orgorgupdate"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := amtest.CreateOrgInstance(orgName)

	if _, _, _, _, err := service.Create(ctx, userContext, org); err != nil {
		t.Fatalf("error creating organization: %s\n", err)
	}
	defer amtest.DeleteOrg(db, orgName, t)

	oid, returned, err := service.Get(ctx, userContext, orgName)
	if err != nil {
		t.Fatalf("error getting organization: %s\n", err)
	}

	updated := &am.Organization{
		StatusID: 2,
	}
	// update usercontext with real orgid
	userContext = testUserContext(oid, 0)
	if _, err := service.Update(ctx, userContext, updated); err != nil {
		t.Fatalf("error updating organization: %s\n", err)
	}

	_, updated, err = service.Get(ctx, userContext, orgName)
	if err != nil {
		t.Fatalf("error getting updated org: %s\n", err)
	}

	if updated.StatusID != 2 {
		t.Fatalf("did not get expected status (2) after update: %d\n", updated.StatusID)
	}

	// manually set returned to updated status so we can compare:
	returned.StatusID = updated.StatusID
	amtest.TestCompareOrganizations(returned, updated, t)

	// ensure we don't change the orgname in an update
	orgNoNameChange := amtest.CreateOrgInstance("orgnonamechange")
	orgNoNameChange.UserPoolID = "newvalue"
	orgNoNameChange.IdentityPoolID = "newvalue"
	orgNoNameChange.FirstName = "newvalue"
	orgNoNameChange.LastName = "newvalue"
	orgNoNameChange.Phone = "newvalue"
	orgNoNameChange.Country = "newvalue"
	orgNoNameChange.City = "newvalue"
	orgNoNameChange.StatePrefecture = "newvalue"
	orgNoNameChange.PostalCode = "newvalue"
	orgNoNameChange.Street = "newvalue"
	orgNoNameChange.StatusID = 1
	orgNoNameChange.SubscriptionID = 1

	if _, err := service.Update(ctx, userContext, orgNoNameChange); err != nil {
		t.Fatalf("error updating organization: %s\n", err)
	}

	if _, _, err := service.Get(ctx, userContext, "orgnonamechange"); err == nil {
		t.Fatalf("error we got back an updated name change: %s\n", err)
	}

	_, newupdated, err := service.Get(ctx, userContext, orgName)
	if err != nil {
		t.Fatalf("error getting updated org back: %s\n", err)
	}
	// manually update orgname so we can compare
	orgNoNameChange.OrgName = orgName
	amtest.TestCompareOrganizations(orgNoNameChange, newupdated, t)
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