package main

import (
	"context"
	"net"
	"os"
	"time"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	moduleservice "github.com/linkai-io/am/protocservices/module"
	modulerprotoc "github.com/linkai-io/am/services/module/protoc"
	"github.com/linkai-io/am/services/module/web"
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

// main starts the WebModuleService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "WebModuleService").Logger()

	sec := secrets.NewSecretsCache(env, region)
	loadBalancerAddr, err = sec.LoadBalancerAddr()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get load balancer address")
	}

	ctx := context.Background()
	browsers := browser.NewGCDBrowserPool(5)
	if err := browsers.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed initializing browsers")
	}
	defer browsers.Close(ctx)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initializers.State(env, region)
	dc := dnsclient.New([]string{"unbound:53"}, 3)

	webDataClient := initializers.WebDataClient(loadBalancerAddr)

	store := filestorage.NewStorage(env, region)
	service := web.New(browsers, webDataClient, dc, state, store)
	err = retrier.Retry(func() error {
		return service.Init()
	})
	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	nsmodulerp := modulerprotoc.New(service)
	moduleservice.RegisterModuleServer(s, nsmodulerp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
