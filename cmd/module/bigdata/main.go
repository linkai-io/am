package main

import (
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/pkg/initializers"

	"github.com/linkai-io/am/mock"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"

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

var region string
var env string
var loadBalancerAddr string

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the BigDataModuleService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "BigDataModuleService").Logger()

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
	dc := dnsclient.New([]string{"unbound:53"}, 3)
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

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
