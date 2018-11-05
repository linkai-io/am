package main

import (
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	addressprotoservice "github.com/linkai-io/am/protocservices/address"
	"github.com/linkai-io/am/services/address"
	addressprotoc "github.com/linkai-io/am/services/address/protoc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var region string
var env string

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the Address Service
func main() {
	var service *address.Service
	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "AddressService").Logger()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	dbstring, db := initializers.DB(env, region, am.AddressServiceKey)

	err = retrier.Retry(func() error {
		policyManager := ladonauth.NewPolicyManager(db, "pgx")
		if err := policyManager.Init(); err != nil {
			log.Fatal().Err(err).Msg("initializing policyManager failed")
		}

		roleManager := ladonauth.NewRoleManager(db, "pgx")
		if err := roleManager.Init(); err != nil {
			log.Fatal().Err(err).Msg("initializing roleManager failed")
		}

		authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)

		service = address.New(authorizer)

		return service.Init([]byte(dbstring))
	})

	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	addressp := addressprotoc.New(service)
	addressprotoservice.RegisterAddressServer(s, addressp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
