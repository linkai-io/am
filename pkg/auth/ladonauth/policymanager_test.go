package ladonauth_test

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/ory/ladon"
	uuid "github.com/satori/go.uuid"

	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/amtest"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
)

func TestNewPolicy(t *testing.T) {
	db := amtest.InitDB(t)
	manager := ladonauth.NewPolicyManager(db, "pgx")
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}
}

func TestCreate(t *testing.T) {
	db := amtest.InitDB(t)
	manager := ladonauth.NewPolicyManager(db, "pgx")
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}
	id := "1"
	expected := &ladon.DefaultPolicy{
		ID:         id,
		Subjects:   []string{"123"},
		Actions:    []string{"create", "update"},
		Effect:     ladon.AllowAccess,
		Resources:  []string{"articles:<[0-9]+>"},
		Conditions: ladon.Conditions{},
	}

	// Test Create/Get
	if err := manager.Create(expected); err != nil {
		t.Fatalf("error creating policy: %s\n", err)
	}

	returned, err := manager.Get(id)
	if err != nil {
		t.Fatalf("error getting policy: %s\n", err)
	}

	testPolicyMatch(expected, returned, t)

	// Test Delete
	if err := manager.Delete(id); err != nil {
		t.Fatalf("error deleting policy: %s\n", err)
	}

	returned, err = manager.Get(id)

	if err == nil {
		t.Fatalf("did not get error when requesting deleted policy\n")
	}

	if returned != nil {
		t.Fatalf("policy should be nil, got %#v\n", returned)
	}

	// Test Update
	if err := manager.Create(expected); err != nil {
		t.Fatalf("error re-creating policy: %s\n", err)
	}

	expected.Subjects = []string{"123", "456"}
	if err := manager.Update(expected); err != nil {
		t.Fatalf("error updating policy: %s\n", err)
	}

	returned, err = manager.Get(id)
	if err != nil {
		t.Fatalf("gott error when requesting updated policy: %s\n", err)
	}

	testPolicyMatch(expected, returned, t)

	// Test Deleting updated policy
	if err := manager.Delete(id); err != nil {
		t.Fatalf("error deleting policy: %s\n", err)
	}

	returned, err = manager.Get(id)
	if err == nil {
		t.Fatalf("did not get error when requesting deleted policy\n")
	}
}

func TestCreateDefaultPolicy(t *testing.T) {
	db := amtest.InitDB(t)
	manager := ladonauth.NewPolicyManager(db, "pgx")
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}

	policy := testCreatePolicy(
		"Manage Custom Tags access to Tag Service",
		[]byte("{\"key\":\"TestManageTagServiceCustom\"}"),
		//subjects
		[]string{am.ReviewerRole},
		//actions
		[]string{"create", "read", "update", "delete"},
		[]string{am.RNTagServiceCustom},
	)

	if err := manager.Create(policy); err != nil {
		t.Fatalf("error creating test default policy: %s\n", err)
	}
	if err := manager.Delete(policy.GetID()); err != nil {
		t.Fatalf("error deleting policy: %s\n", err)
	}
}

func testCreatePolicy(description string, meta []byte, subjects, actions, resources []string) ladon.Policy {
	id, err := uuid.NewV4()
	if err != nil {
		panic(err) // should never happen really
	}

	return &ladon.DefaultPolicy{
		ID:          id.String(),
		Description: description,
		Meta:        meta,
		Subjects:    subjects,
		Actions:     actions,
		Resources:   resources,
		Effect:      ladon.AllowAccess,
		Conditions:  ladon.Conditions{},
	}
}

const defaultPolicyCount = 7

func TestGetAll(t *testing.T) {
	start := 10
	end := 60
	db := amtest.InitDB(t)
	manager := ladonauth.NewPolicyManager(db, "pgx")
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}
	for i := start; i < end; i++ {
		id := fmt.Sprintf("%d", i)
		expected := &ladon.DefaultPolicy{
			ID:         id,
			Subjects:   []string{fmt.Sprintf("%d_123", i)},
			Actions:    []string{"create", "update"},
			Effect:     ladon.AllowAccess,
			Resources:  []string{"articles:<[0-9]+>"},
			Conditions: ladon.Conditions{},
		}

		if err := manager.Create(expected); err != nil {
			t.Fatalf("error creating policy: %s\n", err)
		}
	}
	defer deletePolicies(start, end, manager)

	policies, err := manager.GetAll(10, 0)
	if err != nil {
		t.Fatalf("error getting 10: %s\n", err)
	}
	if len(policies) != 10 {
		t.Fatalf("expected 10 policies, got: %d\n", len(policies))
	}

	policies, err = manager.GetAll(50, 10)
	if err != nil {
		t.Fatalf("error getting from offset 10: %s\n", err)
	}
	if len(policies) != 40+defaultPolicyCount {
		t.Fatalf("expected 40+defaultPolicyCount policies, got: %d\n", len(policies))
	}
}

func TestFind(t *testing.T) {
	start := 8
	end := 12
	db := amtest.InitDB(t)
	manager := ladonauth.NewPolicyManager(db, "pgx")
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}

	for i := start; i < end; i++ {
		expected := &ladon.DefaultPolicy{
			ID:         fmt.Sprintf("%d", i),
			Subjects:   []string{"123"},
			Actions:    []string{"create", "update"},
			Effect:     ladon.AllowAccess,
			Resources:  []string{"articles:<[0-2]+>"},
			Conditions: ladon.Conditions{},
		}
		if err := manager.Create(expected); err != nil {
			t.Fatalf("error creating policy: %s\n", err)
		}
	}
	defer deletePolicies(start, end, manager)

	policies, err := manager.FindRequestCandidates(&ladon.Request{
		Subject:  "f",
		Action:   "asdf",
		Resource: "x",
	})

	if err != nil {
		t.Fatalf("error finding policies: %s\n", err)
	}

	if len(policies) != 0 {
		t.Fatalf("expected 0 policies got: %d\n", len(policies))
	}

	policies, err = manager.FindRequestCandidates(&ladon.Request{
		Subject:  "123",
		Action:   "create",
		Resource: "articles:10",
	})

	if err != nil {
		t.Fatalf("error finding policies: %s\n", err)
	}

	if len(policies) != 4 {
		t.Fatalf("expected 4 policies got: %d\n", len(policies))
	}
}

func deletePolicies(start, end int, manager *ladonauth.LadonPolicyManager) {
	for i := start; i < end; i++ {
		id := fmt.Sprintf("%d", i)
		manager.Delete(id)
	}
}
func testPolicyMatch(expected, returned ladon.Policy, t *testing.T) {

	if expected.GetID() != returned.GetID() {
		t.Fatalf("id does not match %s != %s\n", expected.GetID(), returned.GetID())
	}

	if expected.AllowAccess() != returned.AllowAccess() {
		t.Fatalf("allowaccess does not match %t != %t\n", expected.AllowAccess(), returned.AllowAccess())
	}

	if !testSortEqual(expected.GetSubjects(), returned.GetSubjects(), t) {
		t.Fatalf("subjects do not match: %#v != %#v\n", expected.GetSubjects(), returned.GetSubjects())
	}

	if !testSortEqual(expected.GetActions(), returned.GetActions(), t) {
		t.Fatalf("actions do not match: %#v != %#v\n", expected.GetActions(), returned.GetActions())
	}

	if !testSortEqual(expected.GetResources(), returned.GetResources(), t) {
		t.Fatalf("resources do not match: %#v != %#v\n", expected.GetResources(), returned.GetResources())
	}

	if !reflect.DeepEqual(expected.GetConditions(), returned.GetConditions()) {
		t.Fatalf("conditions do not match: %#v != %#v\n", expected.GetConditions(), returned.GetConditions())
	}
}

func testSortEqual(expected, returned []string, t *testing.T) bool {
	if len(expected) != len(returned) {
		t.Fatalf("slice did not match size: %#v, %#v\n", expected, returned)
	}
	expected_copy := make([]string, len(expected))
	returned_copy := make([]string, len(returned))

	copy(expected_copy, expected)
	copy(returned_copy, returned)

	sort.Strings(expected_copy)
	sort.Strings(returned_copy)

	return reflect.DeepEqual(expected_copy, returned_copy)
}
