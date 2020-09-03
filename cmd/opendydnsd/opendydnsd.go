package main

import (
	"github.com/creekorful/open-dydns/internal/opendydnsd"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if err := opendydnsd.GetApp().Run(os.Args); err != nil {
		log.Err(err).Msg("unable to start OpenDyDNSD.")
		os.Exit(1)
	}
}
