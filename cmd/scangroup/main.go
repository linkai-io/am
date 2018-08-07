package main

import (
	"log"
	"net"
	"os"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
	"gopkg.linkai.io/v1/repos/am/pkg/secrets"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	scangroupprotoservice "gopkg.linkai.io/v1/repos/am/protocservices/scangroup"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
	scangroupprotoc "gopkg.linkai.io/v1/repos/am/services/scangroup/protoc"
)

var region string
var env string

const serviceKey = "scangroupservice"

func init() {
	region = os.Getenv("REGION")
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
	log.Printf("Starting ScanGroup Service\n")

	service := scangroup.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	s := grpc.NewServer()
	sgp := scangroupprotoc.New(service)
	scangroupprotoservice.RegisterScanGroupServer(s, sgp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
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

	p, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 5,
	})

	return dbstring, p
}
