package opendydnsd

import (
	"github.com/creekorful/open-dydns/internal/opendydnsd/api"
	"github.com/creekorful/open-dydns/internal/opendydnsd/daemon"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
)

func GetApp() *cli.App {
	return &cli.App{
		Name:    "opendydnsd",
		Version: "0.1.0",
		Action:  execute,
	}
}

func execute(c *cli.Context) error {
	// Parse the configuration file
	// TODO use cli parameter to get config.yaml
	config, err := daemon.ParseConfig("config.yaml")
	if err != nil {
		log.Err(err).Msg("unable to parse config file")
		return err
	}

	// Configure the logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(zerolog.DebugLevel) // TODO set logging level

	// Display version etc...
	log.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNSD")

	// Instantiate the Daemon
	d, err := daemon.NewDaemon(config)
	if err != nil {
		log.Err(err).Msg("unable to instantiate the daemon")
		return err
	}

	// Instantiate the API
	a, err := api.NewAPI(d, config.SigningKey)
	if err != nil {
		log.Err(err).Msg("unable to instantiate the API")
		return err
	}

	log.Info().Str("Addr", config.ListenAddr).Msg("OpenDyDNSD API started.")
	return a.Start(config.ListenAddr)
}
