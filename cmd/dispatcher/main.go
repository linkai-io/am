package main

import (
	"log"
	"net"
	"os"
	"time"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/address"
	"github.com/linkai-io/am/clients/coordinator"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"
	dispatcherprotoservice "github.com/linkai-io/am/protocservices/dispatcher"
	"github.com/linkai-io/am/services/dispatcher"
	dispatcherprotoc "github.com/linkai-io/am/services/dispatcher/protoc"
	"github.com/linkai-io/am/services/module/ns"
	nsstate "github.com/linkai-io/am/services/module/ns/state"
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
	addrClient := initAddrClient()
	coordClient := initCoordClient()
	modules := initModules(state)

	service := dispatcher.New(addrClient, coordClient, modules, state)
	if err := service.Init(nil); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	log.Printf("Starting Dispatcher Service\n")

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	dispatcherp := dispatcherprotoc.New(service)
	dispatcherprotoservice.RegisterDispatcherServer(s, dispatcherp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initAddrClient() am.AddressService {
	addrClient := address.New()

	err := retrier.RetryUntil(
		func() error {
			return addrClient.Init([]byte(loadBalancerAddr))
		}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatalf("unable to connect to address client: %s\n", err)
	}
	return addrClient
}

func initCoordClient() am.CoordinatorService {
	coordClient := coordinator.New()

	err := retrier.RetryUntil(
		func() error {
			return coordClient.Init([]byte(loadBalancerAddr))
		}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatalf("unable to connect to coordinator client: %s\n", err)
	}
	return coordClient
}

func initModules(state nsstate.Stater) map[am.ModuleType]am.ModuleService {
	nsClient := ns.New(state)

	err := retrier.RetryUntil(
		func() error {
			return nsClient.Init([]byte(loadBalancerAddr))
		}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatalf("unable to connect to ns module client: %s\n", err)
	}

	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = nsClient
	return modules
}

func initState() *redis.State {
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
