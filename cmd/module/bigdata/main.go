package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/pkg/bq"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/lb/consul"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/protocservices/metrics"
	moduleservice "github.com/linkai-io/am/protocservices/module"
	"github.com/linkai-io/am/services/module/bigdata"
	moduleprotoc "github.com/linkai-io/am/services/module/protoc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const (
	serviceKey = am.BigDataModuleServiceKey
)

var (
	appConfig initializers.AppConfig
	bqConfig  bq.ClientConfig
)

func init() {
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	appConfig.ServiceKey = serviceKey
	consulAddr := initializers.ServiceDiscovery(&appConfig)
	consul.RegisterDefault(time.Second*5, consulAddr) // Address comes from CONSUL_HTTP_ADDR or from aws metadata

	// configure bigquery details, credentials come from secretscache.
	bqConfig.DatasetName = os.Getenv("APP_BQ_DATASET_NAME")
	bqConfig.ProjectID = os.Getenv("APP_BQ_PROJECT_ID")
	bqConfig.TableName = os.Getenv("APP_BQ_TABLENAME")
}

// main starts the BigDataModuleService
func main() {
	var err error

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "BigDataModuleService").Logger()

	sec := secrets.NewSecretsCache(appConfig.Env, appConfig.Region)

	bqCredentials, err := sec.BigQueryCredentials()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get bigquery credentials")
	}

	dnsAddrs, err := sec.DNSAddresses()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get dns server addresses")
	}

	log.Info().Str("consul", os.Getenv("CONSUL_HTTP_ADDR")).Strs("dns_servers", dnsAddrs).Msg("initializing...")

	if appConfig.Addr == "" {
		appConfig.Addr = ":50051"
	}

	listener, err := net.Listen("tcp", appConfig.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	state := initializers.State(&appConfig)
	dc := dnsclient.New(dnsAddrs, 3)
	bdService := initializers.BigDataClient()

	bqClient := initializers.BigQueryClient(&bqConfig, []byte(bqCredentials))

	service := bigdata.New(dc, state, bdService, bqClient)
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
