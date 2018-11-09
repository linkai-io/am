package main

import (
	"net"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	lbpb "github.com/bsm/grpclb/grpclb_backend_v1"
	"github.com/bsm/grpclb/load"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/retrier"
	userprotoservice "github.com/linkai-io/am/protocservices/user"
	"github.com/linkai-io/am/services/user"
	userprotoc "github.com/linkai-io/am/services/user/protoc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	region string
	env    string
)

func init() {
	region = os.Getenv("APP_REGION")
	env = os.Getenv("APP_ENV")
}

// main starts the UserService
func main() {
	var service *user.Service

	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "UserService").Logger()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	dbstring, db := initializers.DB(env, region, am.UserServiceKey)

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

		service = user.New(authorizer)

		return service.Init([]byte(dbstring))
	})

	if err != nil {
		log.Fatal().Err(err).Msg("initializing service failed")
	}

	s := grpc.NewServer()
	r := load.NewRateReporter(time.Minute)

	userp := userprotoc.New(service, r)
	userprotoservice.RegisterUserServiceServer(s, userp)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	lbpb.RegisterLoadReportServer(s, r)

	log.Info().Msg("Starting Service")
	if err := s.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
