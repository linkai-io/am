package organization_test

import (
	"context"
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/services/organization"
)

var env string
var dbstring string

const serviceKey = "orgservice"

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

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()
	service := organization.New(roleManager, auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "orgorgcreate"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()

	db := amtest.InitDB(env, t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := &am.Organization{}

	id := uuid.New()

	_, _, _, _, err := service.Create(ctx, userContext, org, id.String())
	if err == nil {
		t.Fatalf("did not get error creating invalid organization\n")
	}

	org = amtest.CreateOrgInstance(orgName)
	defer amtest.DeleteOrg(db, orgName, t)

	id = uuid.New()
	_, _, ocid, ucid, err := service.Create(ctx, userContext, org, id.String())
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

	_, returned, err = service.GetByAppClientID(ctx, userContext, org.UserPoolAppClientID)
	if err != nil {
		t.Fatalf("error getting organization by app clientid: %s\n", err)
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
	filterOrg := &am.Organization{}

	for _, org := range orgs {
		if org.OrgID == returned.OrgID {
			filterOrg = org
		}
	}
	amtest.TestCompareOrganizations(org, filterOrg, t)
}

func TestCreateRoleFail(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "orgcreaterolefail"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()
	roleManager.CreateRoleFn = func(role *am.Role) (string, error) {
		return "", errors.New("unable to add role")
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	// create an invalid context, which causes roles to not be added (and an error should occur)
	userContext := testUserContext(0, 0)
	org := amtest.CreateOrgInstance(orgName)

	id := uuid.New()
	_, _, _, _, err := service.Create(ctx, userContext, org, id.String())
	if err == nil {
		t.Fatalf("role manager did not throw error\n")
	}

	// if we can still get the org, but the role wasn't added, that means it's busted
	// most likely a permission error when deleting.
	_, _, err = service.Get(ctx, userContext, orgName)
	if err == nil {
		t.Fatalf("error role manager failure did not cause org to be deleted")
	}

	defer amtest.DeleteOrg(db, orgName, t)
}

func TestDelete(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "orgorgdelete"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()

	db := amtest.InitDB(env, t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := amtest.CreateOrgInstance(orgName)

	id := uuid.New()

	if _, _, _, _, err := service.Create(ctx, userContext, org, id.String()); err != nil {
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
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "orgorgupdate"

	auth := amtest.MockAuthorizer()
	roleManager := amtest.MockRoleManager()

	db := amtest.InitDB(env, t)
	defer db.Close()

	service := organization.New(roleManager, auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := amtest.CreateOrgInstance(orgName)

	id := uuid.New()

	if _, _, _, _, err := service.Create(ctx, userContext, org, id.String()); err != nil {
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
	orgNoNameChange.UserPoolAppClientID = "newvalue"
	orgNoNameChange.UserPoolJWK = "newvalue"
	orgNoNameChange.UserPoolAppClientSecret = "newvalue"
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
	orgNoNameChange.LimitTLD = 1
	orgNoNameChange.LimitTLDReached = true
	orgNoNameChange.LimitHosts = 25
	orgNoNameChange.LimitHostsReached = true
	orgNoNameChange.LimitCustomWebFlows = 1
	orgNoNameChange.LimitCustomWebFlowsReached = true

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
