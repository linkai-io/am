package main

import (
	"net"
	"os"
	"strings"
	"time"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/address"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"
	moduleservice "github.com/linkai-io/am/protocservices/module"
	"github.com/linkai-io/am/services/module/brute"
	brutemodulerprotoc "github.com/linkai-io/am/services/module/brute/protoc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

// main starts the BruteModuleService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "BruteModuleService").Logger()

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
	dc := dnsclient.New([]string{"unbound:53"}, 1)
	service := brute.New(dc, state)
	err = retrier.Retry(func() error {
		return service.Init(strings.NewReader(list_ten))
	})
	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	brutemodulerp := brutemodulerprotoc.New(service)
	moduleservice.RegisterModuleServer(s, brutemodulerp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	log.Info().Msg("Starting Service")
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
		log.Fatal().Err(err).Msg("error connecting to address server")
	}
	return addrClient
}

func initState() *redis.State {
	redisState := redis.New()
	sec := secrets.NewDBSecrets(env, region)
	cacheConfig, err := sec.CacheConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get cache connection string")
	}

	err = retrier.RetryUntil(func() error {
		return redisState.Init([]byte(cacheConfig))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to redis")
	}
	return redisState
}
