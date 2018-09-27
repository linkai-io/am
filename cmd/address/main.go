package main

import (
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/jackc/pgx"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/secrets"
	addressprotoservice "github.com/linkai-io/am/protocservices/address"
	"github.com/linkai-io/am/services/address"
	addressprotoc "github.com/linkai-io/am/services/address/protoc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var region string
var env string

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the Address Service
func main() {
	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "AddressService").Logger()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	dbstring, db := initDB()

	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	if err := policyManager.Init(); err != nil {
		log.Fatal().Err(err).Msg("initializing policyManager failed")
	}

	roleManager := ladonauth.NewRoleManager(db, "pgx")
	if err := roleManager.Init(); err != nil {
		log.Fatal().Err(err).Msg("initializing roleManager failed")
	}

	authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)

	service := address.New(authorizer)
	if err := service.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	addressp := addressprotoc.New(service)
	addressprotoservice.RegisterAddressServer(s, addressp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}

func initDB() (string, *pgx.ConnPool) {
	sec := secrets.NewDBSecrets(env, region)
	dbstring, err := sec.DBString(am.AddressServiceKey)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get dbstring")
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		log.Fatal().Err(err).Msg("error parsing connection string")
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
		log.Fatal().Err(err).Msg("failed to connect to postgresql")
	}
	return dbstring, p
}
