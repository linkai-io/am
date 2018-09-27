package main

import (
	"flag"
	"net"
	"os"

	"github.com/bsm/grpclb/balancer"
	"github.com/bsm/grpclb/discovery/consul"
	balancerpb "github.com/bsm/grpclb/grpclb_balancer_v1"
	"github.com/hashicorp/consul/api"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var flags struct {
	addr   string
	consul string
}

var region string
var env string

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")

	flag.StringVar(&flags.addr, "addr", ":9999", "Bind address. Default: :9999")
}

func main() {
	flag.Parse()
	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "AMLoadService").Logger()

	log.Info().Msg("Starting Service")
	if err := listenAndServe(); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}

func listenAndServe() error {
	sec := secrets.NewDBSecrets(env, region)
	consulAddr, err := sec.DiscoveryAddr()
	if err != nil || consulAddr == "" {
		log.Fatal().Msg("error getting discovery server address")
	}

	log.Info().Str("discovery_address", consulAddr).Msg("Discovery service found")

	config := api.DefaultConfig()
	config.Address = consulAddr

	discovery, err := consul.New(config)
	if err != nil {
		return err
	}

	lb := balancer.New(discovery, nil)
	defer lb.Reset()

	srv := grpc.NewServer()
	balancerpb.RegisterLoadBalancerServer(srv, lb)

	lis, err := net.Listen("tcp", flags.addr)
	if err != nil {
		return err
	}

	return srv.Serve(lis)
}
