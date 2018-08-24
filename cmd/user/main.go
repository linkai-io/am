package main

import (
	"log"
	"net"
	"os"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/secrets"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	userprotoservice "github.com/linkai-io/am/protocservices/user"
	"github.com/linkai-io/am/services/user"
	userprotoc "github.com/linkai-io/am/services/user/protoc"
)

var (
	region string
	env    string
)

const serviceKey = "userservice"

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
	log.Printf("Starting User Service\n")

	service := user.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	s := grpc.NewServer()
	userp := userprotoc.New(service)
	userprotoservice.RegisterUserServiceServer(s, userp)
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
	if err != nil {
		log.Fatalf("error connecting to db: %s\n", err)
	}

	return dbstring, p
}
