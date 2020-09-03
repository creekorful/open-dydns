package opendydns_cli

import (
	"github.com/creekorful/open-dydns/internal/common"
	"github.com/creekorful/open-dydns/internal/opendydns-cli/config"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func GetApp() *cli.App {
	return &cli.App{
		Name:    "opendydns-cli",
		Usage:   "The OpenDyDNS CLI",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.1.0",
		Flags: []cli.Flag{
			common.GetLogFlag(),
			&cli.StringFlag{
				Name:  "config",
				Value: "opendydns-cli.toml",
			},
		},
		Action: execute,
	}
}

func execute(c *cli.Context) error {
	// Configure log level
	if err := common.ConfigureLogger(c); err != nil {
		return err
	}

	// Load the configuration file
	configFile := c.String("config")
	_, err := config.Load(configFile)
	if err != nil {
		return err
	}

	// Display version etc...
	log.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNS-CLI")

	return nil
}
