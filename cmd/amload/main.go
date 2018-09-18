package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/bsm/grpclb/balancer"
	"github.com/bsm/grpclb/discovery/consul"
	balancerpb "github.com/bsm/grpclb/grpclb_balancer_v1"
	"github.com/hashicorp/consul/api"
	"github.com/linkai-io/am/pkg/secrets"
	"google.golang.org/grpc"
)

var flags struct {
	addr   string
	consul string
}

var region string
var env string

func init() {

}

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")

	flag.StringVar(&flags.addr, "addr", ":8383", "Bind address. Default: :8383")
}

func main() {
	flag.Parse()
	log.Printf("Starting AM Load Balancer Service\n")
	if err := listenAndServe(); err != nil {
		log.Fatal("FATAL", err.Error())
	}
}

func listenAndServe() error {
	sec := secrets.NewDBSecrets(env, region)
	consulAddr, err := sec.DiscoveryAddr()
	if err != nil || consulAddr == "" {
		log.Fatalf("error getting discovery server address\n")
	}

	log.Printf("Discovery service found at: %s\n", consulAddr)

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
