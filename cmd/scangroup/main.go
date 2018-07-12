package main

import (
	"log"
	"net"
	"os"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
	"gopkg.linkai.io/v1/repos/am/services/scangroup/protoc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	protoservice "gopkg.linkai.io/v1/repos/am/protocservices/scangroup"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
)

var dbstring string

func init() {
	dbstring = os.Getenv("TEST_GOOSE_AM_DB_STRING")
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	db := initDB()
	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	roleManager := ladonauth.NewRoleManager(db, "pgx")
	authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)

	service := scangroup.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initialzing service: %s\n", err)
	}

	s := grpc.NewServer()
	sgp := protoc.New(service)
	protoservice.RegisterScanGroupServer(s, sgp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
