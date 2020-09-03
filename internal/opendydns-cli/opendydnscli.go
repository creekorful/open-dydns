package opendydns_cli

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
	"github.com/creekorful/open-dydns/internal/opendydns-cli/config"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

func GetApp() *cli.App {
	return &cli.App{
		Name:    "opendydns-cli",
		Usage:   "The OpenDyDNS CLI",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.1.0",
		Before:  before,
		Flags: []cli.Flag{
			common.GetLogFlag(),
			&cli.StringFlag{
				Name:  "config",
				Value: "opendydns-cli.toml",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "login",
				ArgsUsage: "EMAIL",
				Usage:     "Authenticate against an OpenDyDNS daemon",
				Action:    login,
			},
		},
	}
}

func before(c *cli.Context) error {
	// Configure log level
	if err := common.ConfigureLogger(c); err != nil {
		return err
	}

	// Create configuration file if not exist
	configFile := c.String("config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := config.Save(config.DefaultConfig, configFile); err != nil {
			return err
		}

		return nil
	}

	// Load the configuration file
	_, err := config.Load(configFile)
	if err != nil {
		return err
	}

	// Store configuration file?

	// Display version etc...
	log.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNS-CLI")

	return nil
}

func login(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing EMAIL")
	}

	// TODO check if not already logged in using config file

	// Ask for user password
	fmt.Printf("Password: ")
	_, _ = terminal.ReadPassword(int(os.Stdin.Fd()))

	return nil // TODO implement
}
