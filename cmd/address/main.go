package main

import (
	"log"
	"net"
	"os"
	"time"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/jackc/pgx"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/secrets"
	addressprotoservice "github.com/linkai-io/am/protocservices/address"
	"github.com/linkai-io/am/services/address"
	addressprotoc "github.com/linkai-io/am/services/address/protoc"
	"google.golang.org/grpc"
)

var region string
var env string

const serviceKey = "addressservice"

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

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
	log.Printf("Starting Address Service\n")

	service := address.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	addressp := addressprotoc.New(service)
	addressprotoservice.RegisterAddressServer(s, addressp)
	// Register reflection service on gRPC server.
	//reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initDB() (string, *pgx.ConnPool) {
	sec := secrets.NewDBSecrets(env, region)
	dbstring, err := sec.DBString(serviceKey)
	if err != nil {
		log.Fatalf("unable to get dbstring: %s\n", err)
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		log.Fatalf("error parsing connection string")
	}

	ticker := time.NewTicker(5 * time.Second)
	stopper := time.After(1 * time.Minute)
	defer ticker.Stop()

	var p *pgx.ConnPool
	for {
		select {
		case <-ticker.C:
			p, err = pgx.NewConnPool(pgx.ConnPoolConfig{
				ConnConfig:     conf,
				MaxConnections: 5,
			})
			if err == nil {
				goto READY
			}
			log.Printf("error connecting to db, retrying in 5 seconds...\n")
		case <-stopper:
			p, err = pgx.NewConnPool(pgx.ConnPoolConfig{
				ConnConfig:     conf,
				MaxConnections: 5,
			})
			if err != nil {
				log.Fatalf("error connecting to db after 1 minute: %s\n", err)
			}

		}
	}
READY:
	return dbstring, p
}
