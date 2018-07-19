package user_test

import (
	"context"
	"os"
	"testing"

	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/amtest"
	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/services/user"
)

func TestNew(t *testing.T) {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := user.New(auth)

	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()

	orgName := "usercreate"

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}

	db := amtest.InitDB(t)
	defer db.Close()

	service := user.New(auth)
	if err := service.Init([]byte(os.Getenv("TEST_GOOSE_AM_DB_STRING"))); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	amtest.CreateOrg(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)

	userContext := amtest.CreateUserContext(orgID, 1)
	expected := &am.User{}

	_, err := service.Create(ctx, userContext, expected)
	if err == nil {
		t.Fatalf("did not get error creating invalid user\n")
	}

	expected = testCreateUser(orgName+"testuser@test.local", orgID)

	ucid, err := service.Create(ctx, userContext, expected)
	if err != nil {
		t.Fatalf("error creating organization: %s\n", err)
	}

	if ucid == "" {
		t.Fatalf("invalid ucid returned, was empty\n")
	}

	returned, err := service.GetByCID(ctx, userContext, ucid)
	if err != nil {
		t.Fatalf("error getting user by cid: %s\n", err)
	}

	testCompareUsers(expected, returned, t)

	returned, err = service.Get(ctx, userContext, returned.UserID)
	if err != nil {
		t.Fatalf("error getting user by id: %s\n", err)
	}

	testCompareUsers(expected, returned, t)
}

func testCompareUsers(e, r *am.User, t *testing.T) {

	if e.Deleted != r.Deleted {
		t.Fatalf("Deleted did not match: expected: %v got: %v\n", e.Deleted, r.Deleted)
	}

	if e.FirstName != r.FirstName {
		t.Fatalf("FirstName did not match: expected: %v got: %v\n", e.FirstName, r.FirstName)
	}

	if e.LastName != r.LastName {
		t.Fatalf("LastName did not match: expected: %v got: %v\n", e.LastName, r.LastName)
	}

	if e.UserEmail != r.UserEmail {
		t.Fatalf("UserEmail did not match: expected: %v got: %v\n", e.UserEmail, r.UserEmail)
	}

	if e.OrgID != r.OrgID {
		t.Fatalf("OrgID did not match: expected: %v got: %v\n", e.OrgID, r.OrgID)
	}
}

func testCreateUser(userEmail string, orgID int) *am.User {
	return &am.User{
		OrgID:     orgID,
		UserEmail: userEmail,
		FirstName: "first",
		LastName:  "last",
	}
}
