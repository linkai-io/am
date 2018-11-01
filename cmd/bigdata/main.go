package main

import (
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/jackc/pgx"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	bigdataprotoservice "github.com/linkai-io/am/protocservices/bigdata"
	"github.com/linkai-io/am/services/bigdata"
	bigdataprotoc "github.com/linkai-io/am/services/bigdata/protoc"
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

// main starts the UserService
func main() {
	var service *bigdata.Service

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "BigDataService").Logger()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	dbstring, db := initDB()

	err = retrier.Retry(func() error {
		policyManager := ladonauth.NewPolicyManager(db, "pgx")
		if err := policyManager.Init(); err != nil {
			return errors.Wrap(err, "initializing policyManager failed")
		}

		roleManager := ladonauth.NewRoleManager(db, "pgx")
		if err := roleManager.Init(); err != nil {
			return errors.Wrap(err, "initializing roleManager failed")
		}

		authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)
		log.Info().Msg("Starting Service")
		service = bigdata.New(authorizer)
		return service.Init([]byte(dbstring))
	})

	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	bigdatap := bigdataprotoc.New(service)
	bigdataprotoservice.RegisterBigDataServer(s, bigdatap)
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
	dbstring, err := sec.DBString(am.BigDataServiceKey)
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
