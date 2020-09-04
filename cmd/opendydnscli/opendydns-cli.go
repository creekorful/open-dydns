package main

import (
	"github.com/creekorful/open-dydns/internal/opendydnscli"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if err := opendydnscli.NewCLI().App().Run(os.Args); err != nil {
		log.Err(err).Msg("error while executing OpenDyDNS-CLI.")
		os.Exit(1)
	}
}
