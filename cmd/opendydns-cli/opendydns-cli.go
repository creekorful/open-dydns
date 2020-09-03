package main

import (
	opendydns_cli "github.com/creekorful/open-dydns/internal/opendydns-cli"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if err := opendydns_cli.GetApp().Run(os.Args); err != nil {
		log.Err(err).Msg("unable to start OpenDyDNS-CLI.")
		os.Exit(1)
	}
}
