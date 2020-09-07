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

// DaemonApp represent a instance of the Daemon app
type DaemonApp struct {
	conf     config.Config
	confPath string
	logger   *zerolog.Logger
}

// NewDaemonApp return a new instance of the daemon app
func NewDaemonApp() *DaemonApp {
	return &DaemonApp{}
}

// GetApp return the cli.App representing the DaemonApp
func (da *DaemonApp) GetApp() *cli.App {
	app := &cli.App{
		Name:    "opendydnsd",
		Usage:   "The OpenDyDNS(Daemon)",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.2.0",
		Before:  da.before,
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
				Action:    da.createUser,
			},
		},
		Action: da.startDaemon,
	}

	for _, flag := range common.GetLogFlags() {
		app.Flags = append(app.Flags, flag)
	}

	return app
}

func (da *DaemonApp) before(c *cli.Context) error {
	// Configure log level
	logger, err := common.ConfigureLogger(c)
	if err != nil {
		return err
	}
	da.logger = &logger

	// Create configuration file if not exist
	configFile := c.String("config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		da.logger.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := config.Save(config.DefaultConfig, configFile); err != nil {
			return err
		}
		return fmt.Errorf("please configure the config file")
	}
	da.confPath = configFile

	// Load the configuration file
	conf, err := config.Load(configFile)
	if err != nil {
		return err
	}
	da.conf = conf

	return nil
}

func (da *DaemonApp) startDaemon(c *cli.Context) error {
	// Display version etc...
	da.logger.Info().Str("Version", c.App.Version).Msg("starting OpenDyDNSD")

	// Instantiate the Daemon
	d, err := daemon.NewDaemon(da.conf, da.logger)
	if err != nil {
		da.logger.Err(err).Msg("unable to start the daemon.")
		return err
	}

	// Instantiate the API
	a, err := api.NewAPI(d, da.conf.APIConfig)
	if err != nil {
		da.logger.Err(err).Msg("unable to instantiate the API.")
		return err
	}

	da.logger.Info().Str("Addr", da.conf.APIConfig.ListenAddr).Msg("OpenDyDNSD API started.")
	return a.Start(da.conf.APIConfig.ListenAddr)
}

func (da *DaemonApp) createUser(c *cli.Context) error {
	if c.Args().Len() != 1 {
		err := fmt.Errorf("missing EMAIL")
		da.logger.Err(err).Msg("missing EMAIL.")
		return err
	}

	email := c.Args().First()

	fmt.Printf("Password: ")
	pass, _ := terminal.ReadPassword(int(os.Stdin.Fd()))

	da.logger.Info().Str("Email", email).Msg("creating user.")

	d, err := daemon.NewDaemon(da.conf, da.logger)
	if err != nil {
		da.logger.Err(err).Msg("unable to start the daemon.")
		return err
	}

	if _, err := d.CreateUser(proto.CredentialsDto{
		Email:    email,
		Password: string(pass),
	}); err != nil {
		da.logger.Err(err).Str("Email", email).Msg("unable to create user account.")
		return err
	}

	da.logger.Info().Str("Email", email).Msg("successfully created user account.")

	return nil
}
