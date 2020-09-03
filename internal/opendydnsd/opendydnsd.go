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

var configFile = "config.toml"

func GetApp() *cli.App {
	return &cli.App{
		Name:    "opendydnsd",
		Version: "0.1.0",
		Action:  execute,
	}
}

func execute(c *cli.Context) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(zerolog.DebugLevel)

	// Create configuration file if not exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := config.Save(config.DefaultConfig, configFile); err != nil {
			return err
		}

		return nil
	}

	// Parse the configuration file
	// TODO use cli parameter to get config.toml
	conf, err := config.Load(configFile)
	if err != nil {
		return err
	}

	// Configure the logging
	// TODO set logging level

	// Display version etc...
	log.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNSD")

	// Instantiate the Daemon
	d, err := daemon.NewDaemon(conf)
	if err != nil {
		log.Err(err).Msg("unable to instantiate the daemon")
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
