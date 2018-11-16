package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/pkg/initializers"

	"github.com/linkai-io/am/mock"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	moduleservice "github.com/linkai-io/am/protocservices/module"
	"github.com/linkai-io/am/services/module/bigdata"
	moduleprotoc "github.com/linkai-io/am/services/module/protoc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceKey = am.BigDataModuleServiceKey
)

var (
	appConfig        *initializers.AppConfig
	loadBalancerAddr string
)

func init() {
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	appConfig.ServiceKey = serviceKey
}

// main starts the BigDataModuleService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "BigDataModuleService").Logger()

	sec := secrets.NewSecretsCache(appConfig.Env, appConfig.Region)
	loadBalancerAddr, err = sec.LoadBalancerAddr()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get load balancer address")
	}

	dnsAddrs, err := sec.DNSAddresses()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get dns server addresses")
	}

	if appConfig.Addr == "" {
		appConfig.Addr = ":50051"
	}

	listener, err := net.Listen("tcp", appConfig.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initializers.State(appConfig)
	dc := dnsclient.New(dnsAddrs, 3)
	bdService := &mock.BigDataService{}
	service := bigdata.New(dc, state, bdService, &mock.BigQuerier{})
	err = retrier.Retry(func() error {
		return service.Init(nil)
	})
	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	nsmodulerp := moduleprotoc.New(service, r)
	moduleservice.RegisterModuleServer(s, nsmodulerp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	// check if self register
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initializers.Self(ctx, appConfig)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}