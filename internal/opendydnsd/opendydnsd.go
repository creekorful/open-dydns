package opendydnsd

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
	"github.com/creekorful/open-dydns/internal/opendydnsd/api"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/daemon"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

// OpenDyDNSD represent a instance of the Daemon app
type OpenDyDNSD struct {
	conf     config.Config
	confPath string
	logger   *zerolog.Logger
}

// NewDaemon return a new instance of the daemon app
func NewDaemon() *OpenDyDNSD {
	return &OpenDyDNSD{}
}

// GetApp return the cli.App representing the OpenDyDNSD
func (d *OpenDyDNSD) GetApp() *cli.App {
	app := &cli.App{
		Name:    "opendydnsd",
		Usage:   "The OpenDyDNS(Daemon)",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.1.0",
		Before:  d.before,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "opendydnsd.toml",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "create-user",
				ArgsUsage: "<EMAIL>",
				Usage:     "Create an user account",
				Action:    d.createUser,
			},
		},
		Action: d.startDaemon,
	}

	for _, flag := range common.GetLogFlags() {
		app.Flags = append(app.Flags, flag)
	}

	return app
}

func (d *OpenDyDNSD) before(c *cli.Context) error {
	// Configure log level
	logger, err := common.ConfigureLogger(c)
	if err != nil {
		return err
	}
	d.logger = &logger

	// Create configuration file if not exist
	configFile := c.String("config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		d.logger.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := config.Save(config.DefaultConfig, configFile); err != nil {
			return err
		}
		return fmt.Errorf("please configure the config file")
	}
	d.confPath = configFile

	// Load the configuration file
	conf, err := config.Load(configFile)
	if err != nil {
		return err
	}
	d.conf = conf

	return nil
}

func (d *OpenDyDNSD) startDaemon(c *cli.Context) error {
	// Display version etc...
	d.logger.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNSD")

	// Instantiate the Daemon
	daem, err := daemon.NewDaemon(d.conf, d.logger)
	if err != nil {
		d.logger.Err(err).Msg("Unable to start the daemon.")
		return err
	}

	// Instantiate the API
	a, err := api.NewAPI(daem, d.conf.APIConfig)
	if err != nil {
		d.logger.Err(err).Msg("unable to instantiate the API.")
		return err
	}

	d.logger.Info().Str("Addr", d.conf.APIConfig.ListenAddr).Msg("OpenDyDNSD API started.")
	return a.Start(d.conf.APIConfig.ListenAddr)
}

func (d *OpenDyDNSD) createUser(c *cli.Context) error {
	if c.Args().Len() != 1 {
		err := fmt.Errorf("missing EMAIL")
		d.logger.Err(err).Msg("missing EMAIL.")
		return err
	}

	email := c.Args().First()

	fmt.Printf("Password: ")
	pass, _ := terminal.ReadPassword(int(os.Stdin.Fd()))

	d.logger.Info().Str("Email", email).Msg("Creating user.")

	daem, err := daemon.NewDaemon(d.conf, d.logger)
	if err != nil {
		d.logger.Err(err).Msg("Unable to start the daemon.")
		return err
	}

	if _, err := daem.CreateUser(proto.CredentialsDto{
		Email:    email,
		Password: string(pass),
	}); err != nil {
		d.logger.Err(err).Str("Email", email).Msg("Unable to create user account.")
		return err
	}

	d.logger.Info().Str("Email", email).Msg("Successfully created user account.")

	return nil
}
