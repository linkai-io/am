package main

import (
	"os"

	"github.com/linkai-io/am/pkg/portscanner"
	"github.com/rs/zerolog/log"
)

var env = os.Getenv("APP_ENV")

func main() {
	service := portscanner.NewService()
	if err := service.Init(env); err != nil {
		log.Fatal().Err(err).Msg("failed to start port scanner")
	}

	if err := service.Serve(); err != nil {
		log.Fatal().Err(err).Msg("error serving portscanner")
	}
}
