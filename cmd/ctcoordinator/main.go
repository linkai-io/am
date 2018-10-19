package main

import (
	"os"
	"time"

	"github.com/jackc/pgx"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
)

var region string
var env string
var loadBalancerAddr string

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the CTCoordinatorService
func main() {
	/*
		var err error

		zerolog.TimeFieldFormat = ""
		log.Logger = log.With().Str("service", "CTCoordinatorService").Logger()

		sec := secrets.NewDBSecrets(env, region)
		loadBalancerAddr, err = sec.LoadBalancerAddr()
		if err != nil {
			log.Fatal().Err(err).Msg("unable to get load balancer address")
		}

		listener, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to listen")
		}

			service := coordinator.New(state, scanGroupClient, systemOrgID, systemUserID)
			err = retrier.Retry(func() error {
				return service.Init([]byte(loadBalancerAddr))
			})
			if err != nil {
				log.Fatal().Err(err).Msg("initializing service failed")
			}

			s := grpc.NewServer()
			r := load.NewRateReporter(time.Minute)

			coordp := coordprotoc.New(service)
			coordinatorprotoservice.RegisterCoordinatorServer(s, coordp)
			// Register reflection service on gRPC server.
			reflection.Register(s)
			lbpb.RegisterLoadReportServer(s, r)

			log.Info().Msg("Starting Service")
			if err := s.Serve(listener); err != nil {
				log.Fatal().Err(err).Msg("failed to serve grpc")
			}
	*/
}

func initDB() (string, *pgx.ConnPool) {
	sec := secrets.NewDBSecrets(env, region)
	dbstring, err := sec.DBString(am.CTCoordinatorServiceKey)
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
