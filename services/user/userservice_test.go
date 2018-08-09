package user_test

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/amtest"
	"gopkg.linkai.io/v1/repos/am/mock"
	"gopkg.linkai.io/v1/repos/am/pkg/secrets"
	"gopkg.linkai.io/v1/repos/am/services/user"
)

var env string
var dbstring string

const serviceKey = "userservice"

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
	service := user.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
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

	db := amtest.InitDB(env, t)
	defer db.Close()

	service := user.New(auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	amtest.CreateOrg(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)

	userContext := amtest.CreateUserContext(orgID, 1)
	expected := &am.User{}

	_, _, _, err := service.Create(ctx, userContext, expected)
	if err == nil {
		t.Fatalf("did not get error creating invalid user\n")
	}

	expected = testCreateUser(orgName+"testuser@test.local", orgID)

	_, _, ucid, err := service.Create(ctx, userContext, expected)
	if err != nil {
		t.Fatalf("error creating organization: %s\n", err)
	}

	if ucid == "" {
		t.Fatalf("invalid ucid returned, was empty\n")
	}

	_, returned, err := service.GetByCID(ctx, userContext, ucid)
	if err != nil {
		t.Fatalf("error getting user by cid: %s\n", err)
	}

	testCompareUsers(expected, returned, t)

	_, returned, err = service.Get(ctx, userContext, returned.UserID)
	if err != nil {
		t.Fatalf("error getting user by id: %s\n", err)
	}

	testCompareUsers(expected, returned, t)

	count := 20
	users := make([]*am.User, count)
	for i := 0; i < count; i++ {
		users[i] = testCreateUser(fmt.Sprintf("%d%s@email.com", i, orgName), orgID)
		users[i].OrgID, users[i].UserID, users[i].UserCID, err = service.Create(ctx, userContext, users[i])
		if err != nil {
			t.Fatalf("error creating user %d: %s\n", i, err)
		}
	}

	filter := &am.UserFilter{Start: 0, Limit: 10}
	_, userList, err := service.List(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("got error listing users: %s\n", err)
	}
	if len(userList) != 10 {
		t.Fatalf("expected 10 users got: %d\n", len(userList))
	}

	for i := 0; i < 5; i++ {
		if _, err := service.Delete(ctx, userContext, userList[i].UserID); err != nil {
			t.Fatalf("error deleting user (%d): %s\n", userList[i].UserID, err)
		}
	}
	filter.WithDeleted = true
	filter.DeletedValue = true
	_, userList, err = service.List(ctx, userContext, filter)
	if err != nil {
		t.Fatalf("got error listing users: %s\n", err)
	}
	if len(userList) != 5 {
		t.Fatalf("expected 5 deleted users users got: %d\n", len(userList))
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	orgName := "userupdate"

	auth := amtest.MockEmptyAuthorizer()

	db := amtest.InitDB(env, t)
	defer db.Close()

	service := user.New(auth)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing organization service: %s\n", err)
	}

	amtest.CreateOrg(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)

	userContext := amtest.CreateUserContext(orgID, 1)
	expected := testCreateUser(orgName+"testuser@test.local", orgID)

	_, _, ucid, err := service.Create(ctx, userContext, expected)
	if err != nil {
		t.Fatalf("error creating organization: %s\n", err)
	}

	if ucid == "" {
		t.Fatalf("invalid ucid returned, was empty\n")
	}

	_, returned, err := service.GetByCID(ctx, userContext, ucid)
	if err != nil {
		t.Fatalf("error getting user by cid: %s\n", err)
	}

	update := &am.User{
		FirstName: "first1",
		LastName:  "last1",
	}

	if _, _, err := service.Update(ctx, userContext, update, returned.UserID); err != nil {
		t.Fatalf("error updating user: %s\n", err)
	}

	_, new, err := service.Get(ctx, userContext, returned.UserID)
	if err != nil {
		t.Fatalf("error getting user after update: %s\n", err)
	}

	if new.FirstName != "first1" {
		t.Fatalf("expected name to be updated to first1 got: %s\n", new.FirstName)
	}

	if new.LastName != "last1" {
		t.Fatalf("expected name to be updated to last1 got: %s\n", new.FirstName)
	}
	//manually update returned names, so we can compare everything else:
	returned.FirstName = new.FirstName
	returned.LastName = new.LastName
	testCompareUsers(returned, new, t)

	// test updating status only:
	if _, _, err := service.Update(ctx, userContext, &am.User{StatusID: 1000}, new.UserID); err != nil {
		t.Fatalf("error updating statusid: %s\n", err)
	}

	_, statusUser, err := service.Get(ctx, userContext, returned.UserID)
	if err != nil {
		t.Fatalf("error getting user after update: %s\n", err)
	}

	if statusUser.FirstName != "first1" {
		t.Fatalf("expected name to be updated to first1 got: %s\n", statusUser.FirstName)
	}

	if statusUser.LastName != "last1" {
		t.Fatalf("expected name to be updated to last1 got: %s\n", statusUser.FirstName)
	}

	if statusUser.StatusID != 1000 {
		t.Fatalf("expected status id to be updated to 1000 got: %d\n", statusUser.StatusID)
	}
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
