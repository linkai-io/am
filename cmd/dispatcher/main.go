package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/lb/consul"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/pkg/retrier"
	dispatcherprotoservice "github.com/linkai-io/am/protocservices/dispatcher"
	"github.com/linkai-io/am/protocservices/metrics"
	"github.com/linkai-io/am/services/dispatcher"
	dispatcherprotoc "github.com/linkai-io/am/services/dispatcher/protoc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	serviceKey = am.DispatcherServiceKey
)

var (
	appConfig initializers.AppConfig
)

func init() {
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	appConfig.ServiceKey = serviceKey
	consulAddr := initializers.ServiceDiscovery(&appConfig)
	consul.RegisterDefault(time.Second*5, consulAddr) // Address comes from CONSUL_HTTP_ADDR or from aws metadata
}

// main starts the DispatcherService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "DispatcherService").Logger()

	if appConfig.Addr == "" {
		appConfig.Addr = ":50051"
	}

	listener, err := net.Listen("tcp", appConfig.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initializers.State(&appConfig)
	sgClient := initializers.SGClient()
	addrClient := initializers.AddrClient()
	eventClient := initializers.EventClient()
	modules := initializers.Modules(state)

	service := dispatcher.New(sgClient, eventClient, addrClient, modules, state)
	err = retrier.Retry(func() error {
		return service.Init(nil)
	})

	if err != nil {
		log.Fatal().Err(err).Msg("error initializing service")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	dispatcherp := dispatcherprotoc.New(service, r)
	dispatcherprotoservice.RegisterDispatcherServer(s, dispatcherp)
	healthgrpc.RegisterHealthServer(s, health.NewServer())
	// Register reflection service on gRPC server.
	reflection.Register(s)
	metrics.RegisterLoadReportServer(s, r)

	// check if self register
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initializers.Self(ctx, &appConfig)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
