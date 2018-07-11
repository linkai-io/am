package amtest

import (
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	uuid "github.com/satori/go.uuid"
)

const (
	CreateOrgStmt = `insert into am.organizations 
	(organization_name, owner_email, first_name, last_name, phone, country, state_prefecture, street, city, postal_code, creation_time, subscription_id)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 1);`
	CreateUserStmt = `insert into am.users (organization_id, email, first_name, last_name) values ($1, $2, $3, $4)`
	DeleteOrgStmt  = "delete from am.organizations where organization_name=$1"
	DeleteUserStmt = "delete from am.users where email=$1"
	GetOrgIDStmt   = "select organization_id from am.organizations where organization_name=$1"
	GetUserIDStmt  = "select user_id from am.users where organization_id=$1 and email=$2"
)

func GenerateID(t *testing.T) string {
	id, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("error generating ID: %s\n", err)
	}
	return id.String()
}

func InitDB(t *testing.T) *pgx.ConnPool {
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

func CreateOrg(p *pgx.ConnPool, name string, t *testing.T) {
	tag, err := p.Exec(CreateOrgStmt, name, name+"email@email.com", "r", "r", "1-111-111-1111", "usa", "ca", "1 fake lane", "sf", "90210", time.Now().UnixNano())
	if err != nil {
		t.Fatalf("error creating organization %s: %s\n", name, err)
	}

	orgID := GetOrgID(p, name, t)

	tag, err = p.Exec(CreateUserStmt, orgID, name+"email@email.com", "r", "r")
	if err != nil {
		t.Fatalf("error creating user for %s, %s\n", name, err)
	}
	t.Logf("%#v %s\n", tag, err)
}

func DeleteOrg(p *pgx.ConnPool, name string, t *testing.T) {
	p.Exec(DeleteUserStmt, name+"email@email.com")
	p.Exec(DeleteOrgStmt, name)
}

func GetOrgID(p *pgx.ConnPool, name string, t *testing.T) int32 {
	var orgID int32
	err := p.QueryRow(GetOrgIDStmt, name).Scan(&orgID)
	if err != nil {
		t.Fatalf("error finding org id for %s: %s\n", name, err)
	}
	return orgID
}

func GetUserId(p *pgx.ConnPool, orgID int32, name string, t *testing.T) int32 {
	var userID int32
	err := p.QueryRow(GetUserIDStmt, orgID, name+"email@email.com").Scan(&userID)
	if err != nil {
		t.Fatalf("error finding user id for %s: %s\n", name, err)
	}
	return userID
}
