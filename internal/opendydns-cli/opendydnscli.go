package opendydns_cli

import (
	"github.com/creekorful/open-dydns/internal/common"
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
				Value: "config.toml",
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

	return nil
}
