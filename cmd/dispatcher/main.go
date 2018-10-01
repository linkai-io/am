package main

import (
	"encoding/json"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/address"
	"github.com/linkai-io/am/clients/coordinator"
	"github.com/linkai-io/am/clients/module"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"
	dispatcherprotoservice "github.com/linkai-io/am/protocservices/dispatcher"
	"github.com/linkai-io/am/services/dispatcher"
	dispatcherprotoc "github.com/linkai-io/am/services/dispatcher/protoc"
	nsstate "github.com/linkai-io/am/services/module/ns/state"
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

// main starts the DispatcherService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "DispatcherService").Logger()

	sec := secrets.NewDBSecrets(env, region)
	loadBalancerAddr, err = sec.LoadBalancerAddr()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get load balancer address")
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initState()
	addrClient := initAddrClient()
	modules := initModules(state)

	service := dispatcher.New(addrClient, modules, state)
	err = retrier.Retry(func() error {
		return service.Init(nil)
	})

	if err != nil {
		log.Fatal().Err(err).Msg("error initializing service")
	}

	log.Info().Msg("Starting Service")

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	dispatcherp := dispatcherprotoc.New(service)
	dispatcherprotoservice.RegisterDispatcherServer(s, dispatcherp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}

func initAddrClient() am.AddressService {
	addrClient := address.New()

	err := retrier.RetryUntil(
		func() error {
			return addrClient.Init([]byte(loadBalancerAddr))
		}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to address client")
	}
	return addrClient
}

func initCoordClient() am.CoordinatorService {
	coordClient := coordinator.New()

	err := retrier.RetryUntil(
		func() error {
			return coordClient.Init([]byte(loadBalancerAddr))
		}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to coordinator client")
	}
	return coordClient
}

func initModules(state nsstate.Stater) map[am.ModuleType]am.ModuleService {
	nsClient := module.New()
	cfg := &module.Config{Addr: loadBalancerAddr, ModuleType: am.NSModule}
	data, _ := json.Marshal(cfg)

	err := retrier.RetryUntil(
		func() error {
			return nsClient.Init(data)
		}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to ns module client")
	}

	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = nsClient
	return modules
}

func initState() *redis.State {
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
