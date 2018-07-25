package organization_test

import (
	"context"
	"os"
	"testing"

	"gopkg.linkai.io/v1/repos/am/amtest"

	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/services/organization"
)

func TestNew(t *testing.T) {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := organization.New(auth)

	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	orgName := "orgorgcreate"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(auth)
	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := &am.Organization{}

	_, _, _, _, err := service.Create(ctx, userContext, org)
	if err == nil {
		t.Fatalf("did not get error creating invalid organization\n")
	}

	org = testCreateOrg(orgName)
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
	testCompareOrganizations(org, returned, t)

	var oid int
	oid, returned, err = service.GetByCID(ctx, userContext, ocid)
	if err != nil {
		t.Fatalf("error getting organization by cid: %s\n", err)
	}
	testCompareOrganizations(org, returned, t)

	_, returned, err = service.GetByID(ctx, userContext, oid)
	if err != nil {
		t.Fatalf("error getting organization by cid: %s\n", err)
	}
	testCompareOrganizations(org, returned, t)

	filter := &am.OrgFilter{
		Start: 0,
		Limit: 10,
	}
	orgs, err := service.List(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("error listing organizations: %s\n", err)
	}

	if len(orgs) != 1 {
		t.Fatalf("expected 1 organization in list, got: %d\n", len(orgs))
	}
	testCompareOrganizations(org, orgs[0], t)
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	orgName := "orgorgdelete"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(auth)
	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := testCreateOrg(orgName)

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

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}

	db := amtest.InitDB(t)
	defer db.Close()

	service := organization.New(auth)
	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	userContext := testUserContext(0, 0)
	org := testCreateOrg(orgName)

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
	testCompareOrganizations(returned, updated, t)

	// ensure we don't change the orgname in an update
	orgNoNameChange := testCreateOrg("orgnonamechange")
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
	testCompareOrganizations(orgNoNameChange, newupdated, t)
}

// testCompareOrganizations does not compare fields that are unknown prior to creation
// time (creation time, org id, orgcid)
func testCompareOrganizations(expected, returned *am.Organization, t *testing.T) {
	e := expected
	r := returned

	if e.OrgName != r.OrgName {
		t.Fatalf("org name did not match expected: %v got %v\n", e.OrgName, r.OrgName)
	}

	if e.OwnerEmail != r.OwnerEmail {
		t.Fatalf("OwnerEmail did not match expected: %v got %v\n", e.OwnerEmail, r.OwnerEmail)
	}

	if e.UserPoolID != r.UserPoolID {
		t.Fatalf("UserPoolID did not match expected: %v got %v\n", e.UserPoolID, r.UserPoolID)
	}

	if e.IdentityPoolID != r.IdentityPoolID {
		t.Fatalf("IdentityPoolID did not match expected: %v got %v\n", e.IdentityPoolID, r.IdentityPoolID)
	}

	if e.FirstName != r.FirstName {
		t.Fatalf("FirstName did not match expected: %v got %v\n", e.FirstName, r.FirstName)
	}

	if e.LastName != r.LastName {
		t.Fatalf("LastName did not match expected: %v got %v\n", e.LastName, r.LastName)
	}

	if e.Phone != r.Phone {
		t.Fatalf("Phone did not match expected: %v got %v\n", e.Phone, r.Phone)
	}

	if e.Country != r.Country {
		t.Fatalf("Country did not match expected: %v got %v\n", e.Country, r.Country)
	}

	if e.StatePrefecture != r.StatePrefecture {
		t.Fatalf("StatePrefecture did not match expected: %v got %v\n", e.StatePrefecture, r.StatePrefecture)
	}

	if e.Street != r.Street {
		t.Fatalf("Street did not match expected: %v got %v\n", e.Street, r.Street)
	}

	if e.Address1 != r.Address1 {
		t.Fatalf("Address1 did not match expected: %v got %v\n", e.Address1, r.Address1)
	}

	if e.Address2 != r.Address2 {
		t.Fatalf("Address2 did not match expected: %v got %v\n", e.Address2, r.Address2)
	}

	if e.City != r.City {
		t.Fatalf("City did not match expected: %v got %v\n", e.City, r.City)
	}

	if e.PostalCode != r.PostalCode {
		t.Fatalf("PostalCode did not match expected: %v got %v\n", e.PostalCode, r.PostalCode)
	}

	if e.StatusID != r.StatusID {
		t.Fatalf("StatusID did not match expected: %v got %v\n", e.StatusID, r.StatusID)
	}

	if e.Deleted != r.Deleted {
		t.Fatalf("Deleted did not match expected: %v got %v\n", e.StatePrefecture, r.StatePrefecture)
	}

	if e.SubscriptionID != r.SubscriptionID {
		t.Fatalf("SubscriptionID did not match expected: %v got %v\n", e.StatePrefecture, r.StatePrefecture)
	}

	if r.CreationTime <= 0 {
		t.Fatalf("creation time of returned was not set\n")
	}

}

func testCreateOrg(orgName string) *am.Organization {
	return &am.Organization{
		OrgName:         orgName,
		OwnerEmail:      orgName + "email@email.com",
		UserPoolID:      "userpool.blah",
		IdentityPoolID:  "identitypool.blah",
		FirstName:       "first",
		LastName:        "last",
		Phone:           "1-111-111-1111",
		Country:         "USA",
		City:            "Beverly Hills",
		StatePrefecture: "CA",
		PostalCode:      "90210",
		Street:          "1 fake lane",
		SubscriptionID:  1000,
		StatusID:        1000,
	}
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
