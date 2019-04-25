package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/webflowclient"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var client *webflowclient.Client

var (
	appConfig initializers.AppConfig
)

func init() {
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	store := filestorage.NewStorage(appConfig.Env, appConfig.Region)
	if err := store.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize storage")
	}
	client = webflowclient.New(store)
}

func HandleLambdaEvent(ctx context.Context, event webflowclient.RequestEvent) (*webflowclient.Results, error) {
	zerolog.TimeFieldFormat = ""
	log.Logger = log.With().Str("service", "WebModuleService").Logger()
	return client.Do(ctx, &event)
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
