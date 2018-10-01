package main

import (
	"net"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/scangroup"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"
	"github.com/rs/zerolog"

	coordinatorprotoservice "github.com/linkai-io/am/protocservices/coordinator"
	"github.com/linkai-io/am/services/coordinator"
	coordprotoc "github.com/linkai-io/am/services/coordinator/protoc"
	"github.com/linkai-io/am/services/coordinator/state"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var region string
var env string
var loadBalancerAddr string

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the CoordinatorService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "CoordinatorService").Logger()

	sec := secrets.NewDBSecrets(env, region)
	loadBalancerAddr, err = sec.LoadBalancerAddr()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get load balancer address")
	}

	systemOrgID, err := sec.SystemOrgID()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get system org id")
	}

	systemUserID, err := sec.SystemUserID()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get system user id")
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initState()
	scanGroupClient := initSGClient()

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
}

func initState() state.Stater {
	redisState := redis.New()
	sec := secrets.NewDBSecrets(env, region)
	cacheConfig, err := sec.CacheConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get cache connection string")
	}

	err = retrier.RetryUntil(func() error {
		log.Info().Msg("attempting to connect to redis")
		return redisState.Init([]byte(cacheConfig))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to redis")
	}
	return redisState
}

func initSGClient() am.ScanGroupService {
	scanGroupClient := scangroup.New()

	err := retrier.RetryUntil(func() error {
		return scanGroupClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to scangroup server")
	}
	return scanGroupClient
}
