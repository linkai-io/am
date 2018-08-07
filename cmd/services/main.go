// import gopkg.linkai.io/v1/repos/am/cmd/services
package main

import (
	"log"
	"net"
	"os"

	"github.com/jackc/pgx"
	"gopkg.linkai.io/v1/repos/am/pkg/auth/ladonauth"
	"gopkg.linkai.io/v1/repos/am/services/organization"
	orgprotoc "gopkg.linkai.io/v1/repos/am/services/organization/protoc"
	"gopkg.linkai.io/v1/repos/am/services/scangroup"
	scangroupprotoc "gopkg.linkai.io/v1/repos/am/services/scangroup/protoc"
	"gopkg.linkai.io/v1/repos/am/services/user"
	userprotoc "gopkg.linkai.io/v1/repos/am/services/user/protoc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	orgprotoservice "gopkg.linkai.io/v1/repos/am/protocservices/organization"
	scangroupprotoservice "gopkg.linkai.io/v1/repos/am/protocservices/scangroup"
	userprotoservice "gopkg.linkai.io/v1/repos/am/protocservices/user"
)

var dbstring string
var serviceType string

func init() {
	serviceType = os.Getenv("SERVICE_TYPE")
	dbstring = os.Getenv("DB_STRING")
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	db := initDB()

	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	if err := policyManager.Init(); err != nil {
		log.Fatalf("initializing policyManager failed: %s\n", err)
	}

	roleManager := ladonauth.NewRoleManager(db, "pgx")
	if err := roleManager.Init(); err != nil {
		log.Fatalf("initializing roleManager failed: %s\n", err)
	}

	authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)
	log.Printf("Starting %s\n", serviceType)
	switch serviceType {
	case "orgservice":
		launchOrganizationService(listener, roleManager, authorizer)
	case "userservice":
		launchUserService(listener, authorizer)
	case "scangroupservice":
		launchScanGroupService(listener, authorizer)
	default:
		log.Fatalf("error unknown service type: %s\n", serviceType)
	}
}

func launchOrganizationService(listener net.Listener, roleManager *ladonauth.LadonRoleManager, authorizer *ladonauth.LadonAuthorizer) {
	service := organization.New(roleManager, authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initialzing service: %s\n", err)
	}

	s := grpc.NewServer()
	orgp := orgprotoc.New(service)
	orgprotoservice.RegisterOrganizationServer(s, orgp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func launchUserService(listener net.Listener, authorizer *ladonauth.LadonAuthorizer) {
	service := user.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initialzing service: %s\n", err)
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

func launchScanGroupService(listener net.Listener, authorizer *ladonauth.LadonAuthorizer) {
	service := scangroup.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatalf("error initialzing service: %s\n", err)
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

func initDB() *pgx.ConnPool {

	if dbstring == "" {
		log.Fatalf("dbstring is not set")
	}
	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		log.Fatalf("error parsing connection string")
	}
	p, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 10,
	})
	if err != nil {
		log.Fatalf("error connecting to db: %s\n", err)
	}

	return p
}
