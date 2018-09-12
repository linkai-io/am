package main

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/scangroup"
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

const serviceKey = "coordinatorservice"

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

func main() {
	listener, err := net.Listen("tcp", ":50050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	state := initState()
	scanGroupClient := initClient()

	service := coordinator.New(state, scanGroupClient)
	if err := service.Init(nil); err != nil {
		log.Fatalf("error initializing service: %s\n", err)
	}

	s := grpc.NewServer()
	coordp := coordprotoc.New(service)
	coordinatorprotoservice.RegisterCoordinatorServer(s, coordp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
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
	log.Printf("cacheconfig: %s\n", cacheConfig)
	ticker := time.NewTicker(5 * time.Second)
	stopper := time.After(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			err := redisState.Init([]byte(cacheConfig))
			if err == nil {
				goto READY
			}
			log.Printf("error connecting to state cache, retrying in 5 seconds... %s\n", err)
		case <-stopper:
			log.Fatalf("error connecting to state cache after 1 minute\n")
		}
	}

READY:
	return redisState
}

func initClient() am.ScanGroupService {
	scanGroupClient := scangroup.New()

	ticker := time.NewTicker(5 * time.Second)
	stopper := time.After(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			if err := scanGroupClient.Init([]byte(":50053")); err == nil {
				goto READY
			}
			log.Printf("error connecting to scangroup service, retrying in 5 seconds...\n")
		case <-stopper:
			log.Fatalf("error connecting to scangroup service after 1 minute\n")
		}
	}
READY:
	return scanGroupClient
}
