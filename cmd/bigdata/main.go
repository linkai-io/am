package main

import (
	"context"
	"net"
	"os"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/retrier"
	bigdataprotoservice "github.com/linkai-io/am/protocservices/bigdata"
	"github.com/linkai-io/am/services/bigdata"
	bigdataprotoc "github.com/linkai-io/am/services/bigdata/protoc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	serviceKey = am.BigDataServiceKey
)

var (
	appConfig initializers.AppConfig
)

func init() {
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	appConfig.ServiceKey = serviceKey
}

// main starts the UserService
func main() {
	var service *bigdata.Service

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "BigDataService").Logger()

	if appConfig.Addr == "" {
		appConfig.Addr = ":50051"
	}

	listener, err := net.Listen("tcp", appConfig.Addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	dbstring, db := initializers.DB(&appConfig)

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
		log.Info().Msg("Starting Service")
		service = bigdata.New(authorizer)
		return service.Init([]byte(dbstring))
	})

	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	bigdatap := bigdataprotoc.New(service, r)
	bigdataprotoservice.RegisterBigDataServer(s, bigdatap)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	// check if self register
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initializers.Self(ctx, &appConfig)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
