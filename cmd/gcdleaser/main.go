package main

import (
	"github.com/linkai-io/am/pkg/browser"
	"github.com/rs/zerolog/log"
)

func main() {
	leaser := browser.NewGcdLeaser()
	log.Info().Msgf("Starting browser gcd leaser on: %s", browser.SOCK)

	if err := leaser.Serve(); err != nil {
		log.Fatal().Err(err).Msg("error serving gcd leaser")
	}
}
