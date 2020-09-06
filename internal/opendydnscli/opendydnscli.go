package opendydnscli

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
	cli2 "github.com/creekorful/open-dydns/internal/opendydnscli/cli"
	"github.com/creekorful/open-dydns/internal/opendydnscli/config"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strconv"
)

// CLIApp represent the opendydns-cli running context
type CLIApp struct {
}

// NewCLIApp instantiate a new CLIApp
func NewCLIApp() *CLIApp {
	return &CLIApp{}
}

// App return the cli.App to execute
func (odc *CLIApp) App() *cli.App {
	app := &cli.App{
		Name:    "opendydns-cli",
		Usage:   "The OpenDyDNS CLI",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "opendydns-cli.toml",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "login",
				ArgsUsage: "<EMAIL>",
				Usage:     "Authenticate against an OpenDyDNS daemon",
				Action:    odc.login,
			},
			{
				Name:      "ls",
				ArgsUsage: "<WHAT>",
				Usage:     "List given resource (aliases, domains). Defaults to aliases",
				Action:    odc.ls,
			},
			{
				Name:      "register",
				ArgsUsage: "<ALIAS>",
				Usage:     "Register an alias",
				Action:    odc.register,
			},
			{
				Name:      "rm",
				ArgsUsage: "<ALIAS>",
				Usage:     "Delete an alias",
				Action:    odc.rm,
			},
			{
				Name:      "set-ip",
				ArgsUsage: "<ALIAS> <IP>",
				Usage:     "Override the IP value for given alias",
				Action:    odc.setIP,
			},
			{
				Name:      "set-synchronize",
				ArgsUsage: "<ALIAS> <STATUS>",
				Usage:     "Enable synchronization for given alias",
				Action:    odc.setSynchronize,
			},
			{
				Name:    "synchronize",
				Aliases: []string{"sync"},
				Usage:   "Synchronize enabled aliases with current IP",
				Action:  odc.synchronize,
			},
		},
	}

	for _, flag := range common.GetLogFlags() {
		app.Flags = append(app.Flags, flag)
	}

	return app
}

func (odc *CLIApp) login(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	if !c.Args().Present() {
		err := fmt.Errorf("missing EMAIL")
		logger.Err(err).Msg("missing EMAIL.")
		return err
	}

	// TODO ask for api address too? (and therefore remove Valid())

	// Ask for user password
	fmt.Printf("Password: ")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	// TODO clear screen after that

	if _, err := app.Authenticate(proto.CredentialsDto{
		Email:    c.Args().First(),
		Password: string(password),
	}); err != nil {
		logger.Err(err).Msg("error while authenticating.")
		return err
	}

	logger.Info().Str("Email", c.Args().First()).Msg("successfully authenticated.")

	return nil
}

func (odc *CLIApp) ls(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	if c.Args().First() == "domain" {
		return odc.lsDomains(app, logger)
	}

	return odc.lsAliases(app, logger)
}

func (odc *CLIApp) lsAliases(c cli2.CLI, logger *zerolog.Logger) error {
	aliases, err := c.GetAliases()
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		logger.Info().Msg("no aliases found.")
		return nil
	}

	for _, alias := range aliases {
		logger.Info().
			Str("Domain", alias.Domain).
			Str("Value", alias.Value).
			Bool("Synchronize", alias.Synchronize).
			Msg("")
	}

	return nil
}

func (odc *CLIApp) lsDomains(c cli2.CLI, logger *zerolog.Logger) error {
	domains, err := c.GetDomains()
	if err != nil {
		return err
	}

	if len(domains) == 0 {
		logger.Info().Msg("no domains configured.")
		return nil
	}

	for _, domain := range domains {
		logger.Info().Str("Domain", domain.Domain).Msg("")
	}

	return nil
}

func (odc *CLIApp) register(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	if !c.Args().Present() {
		err := fmt.Errorf("missing ALIAS")
		logger.Err(err).Msg("missing ALIAS.")
		return err
	}

	name := c.Args().First()

	ip, err := odc.getRemoteIP()
	if err != nil {
		logger.Err(err).Msg("error while getting remote IP.")
		return err
	}

	alias, err := app.RegisterAlias(proto.AliasDto{
		Domain: name,
		Value:  ip,
	})

	if err != nil {
		logger.Err(err).Str("Domain", name).Msg("error while registering alias.")
		return err
	}

	logger.Info().Str("Domain", alias.Domain).Msg("successfully registered alias.")
	return nil
}

func (odc *CLIApp) rm(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	if !c.Args().Present() {
		err := fmt.Errorf("missing ALIAS")
		logger.Err(err).Msg("missing ALIAS.")
		return err
	}

	name := c.Args().First()

	if err := app.DeleteAlias(name); err != nil {
		logger.Err(err).Str("Domain", name).Msg("error while deleting alias.")
		return err
	}

	logger.Info().Str("Domain", name).Msg("successfully deleted alias.")
	return nil
}

func (odc *CLIApp) setIP(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	if c.Args().Len() != 2 {
		err := fmt.Errorf("missing ALIAS IP")
		logger.Err(err).Msg("missing ALIAS IP.")
		return err
	}

	alias := c.Args().First()
	ip := c.Args().Get(1)

	al, err := app.UpdateAlias(proto.AliasDto{
		Domain: alias,
		Value:  ip,
	})

	if err != nil {
		logger.Err(err).
			Str("Domain", alias).
			Str("Value", ip).
			Msg("error while updating alias.")
		return err
	}

	logger.Info().
		Str("Domain", al.Domain).
		Str("Value", al.Value).
		Msg("successfully updated alias.")
	return nil
}

func (odc *CLIApp) setSynchronize(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	if c.Args().Len() != 2 {
		err := fmt.Errorf("missing ALIAS STATUS")
		logger.Err(err).Msg("missing ALIAS STATUS.")
		return err
	}

	status, err := strconv.ParseBool(c.Args().Get(1))
	if err != nil {
		logger.Err(err).Msg("invalid status.")
		return err
	}

	if err := app.SetSynchronize(c.Args().First(), status); err != nil {
		logger.Err(err).
			Str("Domain", c.Args().First()).
			Msg("unable to set synchronize status.")
		return err
	}

	m := logger.Info().Str("Domain", c.Args().First())
	if status {
		m.Msg("enable synchronization.")
	} else {
		m.Msg("disable synchronization.")
	}

	return nil
}

func (odc *CLIApp) synchronize(c *cli.Context) error {
	app, logger, err := getInstance(c)
	if err != nil {
		return err
	}

	ip, err := odc.getRemoteIP()
	if err != nil {
		logger.Err(err).Msg("error while getting remote IP.")
		return err
	}

	return app.Synchronize(ip)
}

func (odc *CLIApp) getRemoteIP() (string, error) {
	c := resty.New()
	r, err := c.R().Get("https://ifconfig.me/ip")
	if err != nil {
		return "", err
	}

	return r.String(), nil
}

// TODO better?
func getInstance(c *cli.Context) (cli2.CLI, *zerolog.Logger, error) {
	// Configure log level
	logger, err := common.ConfigureLogger(c)
	if err != nil {
		return nil, defaultLogger(), err
	}

	// Create configuration file if not exist
	configFile := c.String("config")
	configProvider := config.NewFileProvider(configFile)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logger.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := configProvider.Save(config.DefaultConfig); err != nil {
			logger.Err(err).Msg("error while saving config file.")
			return nil, &logger, err
		}

		return nil, &logger, fmt.Errorf("please edit config file")
	}

	app, err := cli2.NewCLI(configFile, &logger)
	if err != nil {
		return nil, nil, err
	}
	return app, &logger, nil
}

func defaultLogger() *zerolog.Logger {
	l := zerolog.New(zerolog.MultiLevelWriter(zerolog.NewConsoleWriter())).
		With().
		Timestamp().
		Logger()
	return &l
}
