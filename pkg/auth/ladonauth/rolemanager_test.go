package ladonauth_test

import (
	"testing"

	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/amtest"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
)

func TestNew(t *testing.T) {
	db := amtest.InitDB(t)
	//testCreateOrg(db, "role_test", t)
	manager := ladonauth.NewRoleManager(db, "pgx")
	//defer testDeleteOrg(db, "role_test", t)
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}
}

func TestRole(t *testing.T) {
	db := amtest.InitDB(t)

	amtest.CreateOrg(db, "create_test1", t)
	defer amtest.DeleteOrg(db, "create_test1", t)

	orgID1 := amtest.GetOrgID(db, "create_test1", t)

	amtest.CreateOrg(db, "create_test2", t)
	defer amtest.DeleteOrg(db, "create_test2", t)

	orgID2 := amtest.GetOrgID(db, "create_test2", t)

	manager := ladonauth.NewRoleManager(db, "pgx")

	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}

	r1 := &am.Role{OrgID: orgID1, RoleName: am.AdminRole}
	r2 := &am.Role{OrgID: orgID2, RoleName: am.AdminRole}

	roleID1, err := manager.CreateRole(r1)
	if err != nil {
		t.Fatalf("error creating role 1: %s\n", err)
	}
	defer manager.DeleteRole(orgID1, roleID1)

	roleID2, err := manager.CreateRole(r2)
	if err != nil {
		t.Fatalf("error creating role 2: %s\n", err)
	}
	defer manager.DeleteRole(orgID2, roleID2)

	// test role 1
	role1, err := manager.Get(orgID1, roleID1)
	if err != nil {
		t.Fatalf("error getting org 1 role 1: %s\n", err)
	}

	if role1.ID != r1.ID || role1.OrgID != r1.OrgID {
		t.Fatalf("roles did not match expected %#v got: %#v\n", r1, role1)
	}

	// test role 2
	role2, err := manager.Get(orgID2, roleID2)
	if err != nil {
		t.Fatalf("error getting org 2 role 2: %s\n", err)
	}

	if role2.ID != r2.ID || role2.OrgID != r2.OrgID {
		t.Fatalf("roles did not match expected %#v got: %#v\n", r1, role1)
	}

	// test getting role2 for orgid1 should not work
	role1, err = manager.Get(orgID1, roleID2)
	if err == nil {
		t.Fatalf("got role back for mismatching orgID: %#v\n", role1)
	}
	t.Logf("err: %s\n", err)

	if _, err = manager.Get(orgID1, "non-existant-role"); err == nil {
		t.Fatalf("did not get error when accessing non-existent roleid")
	}

	if err = manager.DeleteRole(999999, roleID1); err != nil {
		t.Fatalf("got error when deleting valid role id with invalid org id: %s\n", err)
	}

	// Test List
	roles, err := manager.List(orgID1, 10, 0)
	if err != nil {
		t.Fatalf("error listing roles: %s\n", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected 1 role for orgID1 got: %d\n", len(roles))
	}

	if err := manager.DeleteRole(orgID1, r1.ID); err != nil {
		t.Fatalf("error deleting role for orgID1: %s\n", err)
	}

	roles, err = manager.List(orgID1, 10, 0)
	if err != nil {
		t.Fatalf("got error listing roles for orgID1: %s\n", err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected 0 roles after deleterole for orgID1 got: %d\n", len(roles))
	}

	roles, err = manager.List(orgID2, 10, 0)
	if err != nil {
		t.Fatalf("error listing roles: %s\n", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected 1 role for orgID2 got: %d\n", len(roles))
	}

	if err := manager.DeleteRole(orgID1, r1.ID); err != nil {
		t.Fatalf("error deleting role for orgID1: %s\n", err)
	}

}

func TestMembers(t *testing.T) {
	member1 := "members_test1"
	member2 := "members_test2"

	db := amtest.InitDB(t)

	amtest.CreateOrg(db, member1, t)
	defer amtest.DeleteOrg(db, member1, t)

	orgID1 := amtest.GetOrgID(db, member1, t)

	amtest.CreateOrg(db, member2, t)
	defer amtest.DeleteOrg(db, member2, t)

	orgID2 := amtest.GetOrgID(db, member2, t)

	manager := ladonauth.NewRoleManager(db, "pgx")

	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}

	userID1 := amtest.GetUserId(db, orgID1, member1, t)
	userID2 := amtest.GetUserId(db, orgID2, member2, t)
	if userID1 == 0 || userID2 == 0 {
		t.Fatalf("userID's were not retrieved successfully: %d %d\n", userID1, userID2)
	}

	// test FindByMembers where member has no roles assigned
	if _, err := manager.FindByMember(orgID1, userID1, 9999, 0); err != ladonauth.ErrMemberNotFound {
		t.Fatalf("FindByMembers did not return error where the member had no roles assigned")
	}

	expected1 := &am.Role{OrgID: orgID1, Members: []int{userID1}, RoleName: am.AdminRole}
	expected2 := &am.Role{OrgID: orgID2, Members: []int{userID2}, RoleName: am.AdminRole}

	roleID1, err := manager.CreateRole(expected1)
	if err != nil {
		t.Fatalf("error getting role for orgid: %s\n", err)
	}
	defer manager.DeleteRole(orgID1, expected1.ID)

	roleID2, err := manager.CreateRole(expected2)
	if err != nil {
		t.Fatalf("error getting role for orgid: %s\n", err)
	}
	defer manager.DeleteRole(orgID2, expected2.ID)

	// Test getting org id roles match expected
	r1, err := manager.Get(orgID1, roleID1)
	if err != nil {
		t.Fatalf("error getting role for orgID: %s\n", err)
	}

	if r1.Members[0] != expected1.Members[0] {
		t.Fatalf("error members didn't match, expected %d got %d\n", expected1.Members[0], r1.Members[0])
	}

	r2, err := manager.Get(orgID2, roleID2)
	if err != nil {
		t.Fatalf("error getting role for orgID: %s\n", err)
	}

	if r2.Members[0] != expected2.Members[0] {
		t.Fatalf("error members didn't match, expected %d got %d\n", expected2.Members[0], r2.Members[0])
	}

	// Test FindByMember
	roles, err := manager.FindByMember(orgID1, r1.Members[0], 10, 0)
	if err != nil {
		t.Fatalf("error finding by members: %s\n", err)
	}

	if len(roles) != 1 {
		t.Fatalf("expected 1 role got: %d\n", len(roles))
	}

	if roles[0].ID != r1.ID {
		t.Fatalf("expected findbymember role to match %s got %s\n", r1.ID, roles[0].ID)
	}

	// Test RemoveMembers
	if err := manager.RemoveMembers(orgID1, "non-role", r1.Members); err != nil {
		t.Fatalf("got error removing member non-existent role: %s", err)
	}

	if err := manager.RemoveMembers(orgID1, r1.ID, r1.Members); err != nil {
		t.Fatalf("got error removing member from role: %s", err)
	}

	r, err := manager.Get(orgID1, r1.ID)
	if err != nil {
		t.Fatalf("error getting orgID1's %s role: %s\n", r1.ID, err)
	}

	if r == nil {
		t.Fatalf("role should not be nil")
	}

	if len(r.Members) != 0 {
		t.Fatalf("members should be empty for role: %s\n", r1.ID)
	}
}

func TestGetByName(t *testing.T) {
	name1 := "byname_test1"
	name2 := "byname_test2"

	db := amtest.InitDB(t)

	amtest.CreateOrg(db, name1, t)
	defer amtest.DeleteOrg(db, name1, t)

	orgID1 := amtest.GetOrgID(db, name1, t)

	amtest.CreateOrg(db, name2, t)
	defer amtest.DeleteOrg(db, name2, t)

	orgID2 := amtest.GetOrgID(db, name2, t)

	manager := ladonauth.NewRoleManager(db, "pgx")

	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}

	userID1 := amtest.GetUserId(db, orgID1, name1, t)
	userID2 := amtest.GetUserId(db, orgID2, name2, t)
	if userID1 == 0 || userID2 == 0 {
		t.Fatalf("userID's were not retrieved successfully: %d %d\n", userID1, userID2)
	}

	expected1 := &am.Role{OrgID: orgID1, Members: []int{userID1}, RoleName: am.AdminRole}
	expected2 := &am.Role{OrgID: orgID2, Members: []int{userID2}, RoleName: am.AdminRole}

	roleID1, err := manager.CreateRole(expected1)
	if err != nil {
		t.Fatalf("error getting role for orgid: %s\n", err)
	}
	defer manager.DeleteRole(orgID1, expected1.ID)

	roleID2, err := manager.CreateRole(expected2)
	if err != nil {
		t.Fatalf("error getting role for orgid: %s\n", err)
	}
	defer manager.DeleteRole(orgID2, expected2.ID)

	role1, err := manager.GetByName(orgID1, am.AdminRole)
	if err != nil {
		t.Fatalf("error getting by name for orgID1: %s\n", err)
	}
	role2, err := manager.GetByName(orgID2, am.AdminRole)
	if err != nil {
		t.Fatalf("error getting by name for orgID2: %s\n", err)
	}

	if role1.ID != roleID1 || role2.ID != roleID2 {
		t.Fatalf("returned role id's do not match when getting by name expected: %s and %s got %s and %s\n", role1.ID, roleID1, role2.ID, roleID2)
	}
}
