package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/am"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/jackc/pgx"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	orgprotoservice "github.com/linkai-io/am/protocservices/organization"
	"github.com/linkai-io/am/services/organization"
	orgprotoc "github.com/linkai-io/am/services/organization/protoc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	region string
	env    string
)

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the OrganizationService
func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	dbstring, db := initDB()

	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	if err := policyManager.Init(); err != nil {
		log.Fatalf("initializing policyManager failed: %s\n", err)
	}

	roleManager := ladonauth.NewRoleManager(db, "pgx")
	if err := roleManager.Init(); err != nil {
		log.Fatalf("initializing roleManager failed: %s\n", err)
	}

	authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)
	log.Printf("Starting Organization Service\n")

	service := organization.New(roleManager, authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	orgp := orgprotoc.New(service)
	orgprotoservice.RegisterOrganizationServer(s, orgp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initDB() (string, *pgx.ConnPool) {
	sec := secrets.NewDBSecrets(env, region)
	dbstring, err := sec.DBString(am.OrganizationServiceKey)
	if err != nil {
		log.Fatalf("unable to get dbstring: %s\n", err)
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		log.Fatalf("error parsing connection string: %s\n", err)
	}

	var p *pgx.ConnPool

	err = retrier.RetryUntil(func() error {
		p, err = pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig:     conf,
			MaxConnections: 5,
		})
		return err
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatalf("unable to connect to postgresql: %s\n", err)
	}
	return dbstring, p
}
