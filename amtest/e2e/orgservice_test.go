package e2e_test

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/amtest"
	client "gopkg.linkai.io/v1/repos/am/clients/organization"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
	"gopkg.linkai.io/v1/repos/am/services/organization"
	"gopkg.linkai.io/v1/repos/am/services/organization/protoc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	protoservice "gopkg.linkai.io/v1/repos/am/protocservices/organization"
)

var s *grpc.Server
var orgServerAddr = ":50052"
var dbstring = os.Getenv("ORGSERVICE_DB_STRING")
var roleManager *ladonauth.LadonRoleManager

func TestOrganization(t *testing.T) {
	orgName := "orge2etest"
	db := initDB()
	go organizationServer(db, t)
	time.Sleep(1 * time.Second)
	c := client.New()
	if err := c.Init([]byte(orgServerAddr)); err != nil {
		t.Fatalf("error starting client: %s\n", err)
	}
	ctx := context.Background()
	userContext := &am.UserContextData{
		OrgID:  1,
		UserID: 1,
	}
	defer amtest.DeleteOrg(db, orgName, t)

	org := amtest.CreateOrgInstance(orgName)
	userContext.Roles = []string{am.RNSystem}
	oid, uid, ocid, ucid, err := c.Create(ctx, userContext, org)
	if err != nil {
		t.Fatalf("error creating org: %s\n", err)
	}

	t.Logf("%d %d %s %s\n", oid, uid, ocid, ucid)

	_, returned, err := c.Get(ctx, userContext, orgName)
	if err != nil {
		t.Fatalf("error getting org by name: %s\n", err)
	}

	amtest.TestCompareOrganizations(org, returned, t)

	_, returned, err = c.GetByID(ctx, userContext, oid)
	if err != nil {
		t.Fatalf("error getting org by id: %s\n", err)
	}
	amtest.TestCompareOrganizations(org, returned, t)

	// change user context to the newly created org user
	newUserContext := &am.UserContextData{
		OrgID:  oid,
		UserID: uid,
		OrgCID: ocid,
		Roles:  []string{am.OwnerRole},
	}

	_, returned, err = c.GetByCID(ctx, newUserContext, ocid)
	if err != nil {
		t.Fatalf("error getting org by id with newuser context: %s\n", err)
	}
	amtest.TestCompareOrganizations(org, returned, t)

	// test normal owner can not list orgs
	if _, err = c.List(ctx, newUserContext, &am.OrgFilter{}); err == nil {
		t.Fatalf("normal user allowed to call list")
	}
	t.Logf("%s\n", err)

	// or Get by org name
	if _, _, err = c.Get(ctx, newUserContext, orgName); err == nil {
		t.Fatalf("normal user allowed to call list")
	}

	// or GetByID
	if _, _, err = c.GetByID(ctx, newUserContext, oid); err == nil {
		t.Fatalf("normal user allowed to call list")
	}

	oid, err = c.Delete(ctx, userContext, oid)
	if err != nil {
		t.Fatalf("error deleting org: %s\n", err)
	}

	s.Stop()

}

func organizationServer(db *pgx.ConnPool, t *testing.T) {
	lis, err := net.Listen("tcp", orgServerAddr)
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	authorizer := initAuthorizer(db, t)
	service := organization.New(roleManager, authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initialzing service: %s\n", err)
	}

	s = grpc.NewServer()
	orgp := protoc.New(service)
	protoservice.RegisterOrganizationServer(s, orgp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func initAuthorizer(db *pgx.ConnPool, t *testing.T) *ladonauth.LadonAuthorizer {
	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	if err := policyManager.Init(); err != nil {
		t.Fatalf("error creating policy manager: %s\n", err)
	}
	roleManager = ladonauth.NewRoleManager(db, "pgx")
	if err := roleManager.Init(); err != nil {
		t.Fatalf("error creating role manager: %s\n", err)
	}

	return ladonauth.NewLadonAuthorizer(policyManager, roleManager)
}

func initDB() *pgx.ConnPool {

	if dbstring == "" {
		log.Fatalf("dbstring is not set")
	}
	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		log.Fatalf("error parsing connection string")
	}
	p, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: conf})
	if err != nil {
		log.Fatalf("error connecting to db: %s\n", err)
	}

	return p
}
