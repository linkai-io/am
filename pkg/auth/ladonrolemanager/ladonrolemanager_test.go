package ladonrolemanager_test

import (
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonrolemanager"
)

const (
	testCreateOrgStmt = `insert into am.organizations 
	(organization_name, owner_email, first_name, last_name, phone, country, state_prefecture, street, city, postal_code, creation_time, subscription_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1);`
	testCreateUserStmt = `insert into am.users (organization_id, email, first_name, last_name) values ($1, $2, $3, $4)`
	testDeleteOrgStmt  = "DELETE FROM am.organizations WHERE organization_name=$1"
	testDeleteUserStmt = "DELETE FROM am.users WHERE email=$1"
	testGetOrgIDStmt   = "SELECT organization_id from am.organizations where organization_name=$1"
)

func TestNew(t *testing.T) {
	db := initDB(t)
	//testCreateOrg(db, "role_test", t)
	manager := ladonrolemanager.New(db, "pgx")
	//defer testDeleteOrg(db, "role_test", t)
	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}
}

func TestCreateGetRole(t *testing.T) {
	db := initDB(t)

	testCreateOrg(db, "create_test1", t)
	defer testDeleteOrg(db, "create_test1", t)

	orgID1 := testGetOrgID(db, "create_test1", t)

	testCreateOrg(db, "create_test2", t)
	defer testDeleteOrg(db, "create_test2", t)

	orgID2 := testGetOrgID(db, "create_test2", t)

	manager := ladonrolemanager.New(db, "pgx")

	if err := manager.Init(); err != nil {
		t.Fatalf("error init manager: %s\n", err)
	}

	r1 := &am.Role{OrgID: orgID1}
	r2 := &am.Role{OrgID: orgID2}

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

	// test getting role2 for orgid1
	role1, err = manager.Get(orgID1, roleID2)
	if err == nil {
		t.Fatalf("got role back for mismatching orgID: %#v\n", role1)
	}
}

func initDB(t *testing.T) *pgx.ConnPool {
	dbstring := os.Getenv("TEST_GOOSE_AM_DB_STRING")
	if dbstring == "" {
		t.Fatalf("dbstring is not set")
	}
	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		t.Fatalf("error parsing connection string")
	}
	p, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: conf})
	if err != nil {
		t.Fatalf("error connecting to db: %s\n", err)
	}

	return p
}

func testCreateOrg(p *pgx.ConnPool, name string, t *testing.T) {
	tag, err := p.Exec(testCreateOrgStmt, name, name+"email@email.com", "r", "r", "1-111-111-1111", "usa", "ca", "1 fake lane", "sf", "90210", time.Now().UnixNano())
	if err != nil {
		t.Fatalf("error creating organization %s: %s\n", name, err)
	}

	orgID := testGetOrgID(p, name, t)

	tag, err = p.Exec(testCreateUserStmt, orgID, name+"email@email.com", "r", "r")
	if err != nil {
		t.Fatalf("error creating user for %s, %s\n", name, err)
	}
	t.Logf("%#v %s\n", tag, err)
}

func testDeleteOrg(p *pgx.ConnPool, name string, t *testing.T) {
	p.Exec(testDeleteUserStmt, name+"email@email.com")
	p.Exec(testDeleteOrgStmt, name)
}

func testGetOrgID(p *pgx.ConnPool, name string, t *testing.T) int32 {
	var orgID int32
	err := p.QueryRow(testGetOrgIDStmt, name).Scan(&orgID)
	if err != nil {
		t.Fatalf("error finding org id for %s: %s\n", name, err)
	}
	return orgID
}
