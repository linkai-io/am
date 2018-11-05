package main

import (
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	dispatcherprotoservice "github.com/linkai-io/am/protocservices/dispatcher"
	"github.com/linkai-io/am/services/dispatcher"
	dispatcherprotoc "github.com/linkai-io/am/services/dispatcher/protoc"
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

	sec := secrets.NewSecretsCache(env, region)
	loadBalancerAddr, err = sec.LoadBalancerAddr()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get load balancer address")
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initializers.State(env, region)
	sgClient := initializers.SGClient(loadBalancerAddr)
	addrClient := initializers.AddrClient(loadBalancerAddr)
	modules := initializers.Modules(state, loadBalancerAddr)

	service := dispatcher.New(sgClient, addrClient, modules, state)
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
