package main

import (
	"context"
	"net"
	"os"
	"time"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/retrier"
	scangroupprotoservice "github.com/linkai-io/am/protocservices/scangroup"
	"github.com/linkai-io/am/services/scangroup"
	scangroupprotoc "github.com/linkai-io/am/services/scangroup/protoc"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceKey = am.ScanGroupServiceKey
)

var (
	appConfig *initializers.AppConfig
)

func init() {
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	appConfig.ServiceKey = serviceKey
}

// main starts the ScanGroupService
func main() {
	var service *scangroup.Service

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "ScanGroupService").Logger()

	if appConfig.Addr == "" {
		appConfig.Addr = ":50051"
	}

	listener, err := net.Listen("tcp", appConfig.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	dbstring, db := initializers.DB(appConfig)

	err = retrier.Retry(func() error {
		policyManager := ladonauth.NewPolicyManager(db, "pgx")
		if err := policyManager.Init(); err != nil {
			return errors.Wrap(err, "initializing policyManager failed")
		}

		roleManager := ladonauth.NewRoleManager(db, "pgx")
		if err := roleManager.Init(); err != nil {
			return errors.Wrap(err, "initializing roleManager failed")
		}

		authorizer := ladonauth.NewLadonAuthorizer(policyManager, roleManager)
		service = scangroup.New(authorizer)
		return service.Init([]byte(dbstring))
	})
	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	sgp := scangroupprotoc.New(service, r)
	scangroupprotoservice.RegisterScanGroupServer(s, sgp)
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
