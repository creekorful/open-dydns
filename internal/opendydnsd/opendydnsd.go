package opendydnsd

import (
	"github.com/creekorful/open-dydns/internal/opendydnsd/api"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/daemon"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
)

func GetApp() *cli.App {
	return &cli.App{
		Name:    "opendydnsd",
		Usage:   "The OpenDyDNS(Daemon)",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
			},
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
	// TODO move to common code when implementing the CLI
	lvl, err := zerolog.ParseLevel(c.String("log-level"))
	if err != nil {
		return err
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(lvl)

	// Create configuration file if not exist
	configFile := c.String("config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := config.Save(config.DefaultConfig, configFile); err != nil {
			return err
		}

		return nil
	}

	// Parse the configuration file
	conf, err := config.Load(configFile)
	if err != nil {
		return err
	}

	// Display version etc...
	log.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNSD")

	// Instantiate the Daemon
	d, err := daemon.NewDaemon(conf)
	if err != nil {
		return err
	}

	// Instantiate the API
	a, err := api.NewAPI(d, conf.ApiConfig)
	if err != nil {
		log.Err(err).Msg("unable to instantiate the API")
		return err
	}

	log.Info().Str("Addr", conf.ApiConfig.ListenAddr).Msg("OpenDyDNSD API started.")
	return a.Start(conf.ApiConfig.ListenAddr)
}