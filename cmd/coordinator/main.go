package main

import (
	"log"
	"net"
	"os"
	"time"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/scangroup"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"

	coordinatorprotoservice "github.com/linkai-io/am/protocservices/coordinator"
	"github.com/linkai-io/am/services/coordinator"
	coordprotoc "github.com/linkai-io/am/services/coordinator/protoc"
	"github.com/linkai-io/am/services/coordinator/state"
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

// main starts the CoordinatorService
func main() {
	var err error

	sec := secrets.NewDBSecrets(env, region)
	loadBalancerAddr, err = sec.LoadBalancerAddr()
	if err != nil {
		log.Fatalf("unable to get load balancer address: %s\n", err)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	state := initState()
	scanGroupClient := initSGClient()

	service := coordinator.New(state, scanGroupClient)
	if err := service.Init(nil); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	coordp := coordprotoc.New(service)
	coordinatorprotoservice.RegisterCoordinatorServer(s, coordp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initState() state.Stater {
	redisState := redis.New()
	sec := secrets.NewDBSecrets(env, region)
	cacheConfig, err := sec.CacheConfig()
	if err != nil {
		log.Fatalf("unable to get cache connection string: %s\n", err)
	}

	err = retrier.RetryUntil(func() error {
		return redisState.Init([]byte(cacheConfig))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatalf("error connecting to redis: %s\n", err)
	}
	return redisState
}

func initSGClient() am.ScanGroupService {
	scanGroupClient := scangroup.New()

	err := retrier.RetryUntil(func() error {
		return scanGroupClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatalf("error connecting to scangroup server: %s\n", err)
	}
	return scanGroupClient
}
