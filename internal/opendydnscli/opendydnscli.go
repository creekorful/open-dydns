package opendydnscli

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/common"
	"github.com/creekorful/open-dydns/internal/opendydnscli/client"
	"github.com/creekorful/open-dydns/internal/opendydnscli/config"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strconv"
)

// OpenDYDNSCLI represent the opendydns-cli running context
type OpenDYDNSCLI struct {
	conf     config.Config
	confPath string
	logger   *zerolog.Logger
}

// NewCLI instantiate a new OpenDYDNSCLI
func NewCLI() *OpenDYDNSCLI {
	return &OpenDYDNSCLI{}
}

// App return the cli.App to execute
func (odc *OpenDYDNSCLI) App() *cli.App {
	app := &cli.App{
		Name:    "opendydns-cli",
		Usage:   "The OpenDyDNS CLI",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.1.0",
		Before:  odc.before,
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
				Name:      "add",
				ArgsUsage: "<ALIAS>",
				Usage:     "Register an alias",
				Action:    odc.add,
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

func (odc *OpenDYDNSCLI) before(c *cli.Context) error {
	// Configure log level
	logger, err := common.ConfigureLogger(c)
	if err != nil {
		return err
	}

	odc.logger = &logger

	// Create configuration file if not exist
	configFile := c.String("config")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		odc.logger.Info().Str("Path", configFile).Msg("creating default config file. please edit it accordingly.")
		if err := config.Save(config.DefaultConfig, configFile); err != nil {
			odc.logger.Err(err).Msg("error while saving config file.")
			return err
		}

		return nil
	}

	// Load the configuration file
	conf, err := config.Load(configFile)
	if err != nil {
		odc.logger.Err(err).Msg("error while loading config file.")
		return err
	}

	// Store configuration file
	odc.conf = conf
	odc.confPath = configFile

	return nil
}

func (odc *OpenDYDNSCLI) login(c *cli.Context) error {
	if !c.Args().Present() {
		err := fmt.Errorf("missing EMAIL")
		odc.logger.Err(err).Msg("missing EMAIL.")
		return err
	}

	// check if not already logged in
	if odc.conf.Token != "" {
		err := fmt.Errorf("already logged in")
		odc.logger.Err(err).Msg("already logged in.")
		return err
	}

	// TODO ask for api address too? (and therefore remove Valid())

	// Ask for user password
	fmt.Printf("Password: ")
	password, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
	// TODO clear screen after that

	apiClient := client.NewClient(odc.conf.APIAddr)

	token, err := apiClient.Authenticate(proto.CredentialsDto{
		Email:    c.Args().First(),
		Password: string(password),
	})

	if err != nil {
		odc.logger.Err(err).Msg("error while authenticating.")
		return err
	}

	// Save token in config file
	odc.conf.Token = token.Token
	if err := odc.saveConfig(odc.conf); err != nil {
		odc.logger.Err(err).Msg("error while saving config.")
		return err
	}

	odc.logger.Info().Str("Email", c.Args().First()).Msg("successfully authenticated.")

	return nil
}

func (odc *OpenDYDNSCLI) ls(c *cli.Context) error {
	token, err := odc.getToken()
	if err != nil {
		return err
	}

	if c.Args().First() == "domain" {
		return odc.lsDomains(token)
	}

	return odc.lsAliases(token)
}

func (odc *OpenDYDNSCLI) lsAliases(token proto.TokenDto) error {
	apiClient := client.NewClient(odc.conf.APIAddr)

	aliases, err := apiClient.GetAliases(token)
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		odc.logger.Info().Msg("no aliases found.")
		return nil
	}

	for _, alias := range aliases {
		status := false
		for _, confAlias := range odc.conf.Aliases {
			if confAlias.Name == alias.Domain {
				status = confAlias.Synchronize
			}
		}

		odc.logger.Info().Str("Domain", alias.Domain).Str("Value", alias.Value).Bool("Synchronize", status).Msg("")
	}

	return nil
}

func (odc *OpenDYDNSCLI) lsDomains(token proto.TokenDto) error {
	apiClient := client.NewClient(odc.conf.APIAddr)

	domains, err := apiClient.GetDomains(token)
	if err != nil {
		return err
	}

	if len(domains) == 0 {
		odc.logger.Info().Msg("no domains configured.")
		return nil
	}

	for _, domain := range domains {
		odc.logger.Info().Str("Domain", domain.Domain).Msg("")
	}

	return nil
}

func (odc *OpenDYDNSCLI) add(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing ALIAS")
	}

	name := c.Args().First()

	apiClient := client.NewClient(odc.conf.APIAddr)

	token, err := odc.getToken()
	if err != nil {
		odc.logger.Err(err).Msg("error while getting JWT token.")
		return err
	}

	ip, err := odc.getRemoteIP()
	if err != nil {
		odc.logger.Err(err).Msg("error while getting remote IP.")
		return err
	}

	alias, err := apiClient.RegisterAlias(token, proto.AliasDto{
		Domain: name,
		Value:  ip,
	})

	if err != nil {
		odc.logger.Err(err).Str("Domain", name).Msg("error while registering alias.")
		return err
	}

	odc.logger.Info().Str("Domain", alias.Domain).Msg("successfully created alias.")
	return nil
}

func (odc *OpenDYDNSCLI) rm(c *cli.Context) error {
	if !c.Args().Present() {
		err := fmt.Errorf("missing ALIAS")
		odc.logger.Err(err).Msg("missing ALIAS.")
		return err
	}

	name := c.Args().First()

	apiClient := client.NewClient(odc.conf.APIAddr)

	token, err := odc.getToken()
	if err != nil {
		odc.logger.Err(err).Msg("error while getting JWT token.")
		return err
	}

	if err := apiClient.DeleteAlias(token, name); err != nil {
		odc.logger.Err(err).Str("Domain", name).Msg("error while deleting alias.")
		return err
	}

	odc.logger.Info().Str("Domain", name).Msg("successfully deleted alias.")
	return nil
}

func (odc *OpenDYDNSCLI) setIP(c *cli.Context) error {
	if c.Args().Len() != 2 {
		err := fmt.Errorf("missing ALIAS IP")
		odc.logger.Err(err).Msg("missing ALIAS IP.")
		return err
	}

	alias := c.Args().First()
	ip := c.Args().Get(1)

	apiClient := client.NewClient(odc.conf.APIAddr)

	token, err := odc.getToken()
	if err != nil {
		odc.logger.Err(err).Msg("error while getting JWT token.")
		return err
	}

	al, err := apiClient.UpdateAlias(token, proto.AliasDto{
		Domain: alias,
		Value:  ip,
	})

	if err != nil {
		odc.logger.Err(err).Str("Domain", alias).Str("Value", ip).Msg("error while updating alias.")
		return err
	}

	odc.logger.Info().Str("Domain", al.Domain).Str("Value", al.Value).Msg("successfully updated alias.")
	return nil
}

func (odc *OpenDYDNSCLI) setSynchronize(c *cli.Context) error {
	if c.Args().Len() != 2 {
		err := fmt.Errorf("missing ALIAS STATUS")
		odc.logger.Err(err).Msg("missing ALIAS STATUS.")
		return err
	}

	status, err := strconv.ParseBool(c.Args().Get(1))
	if err != nil {
		odc.logger.Err(err).Msg("invalid status.")
		return err
	}

	conf := odc.conf
	if conf.Aliases == nil {
		conf.Aliases = []config.AliasConfig{}
	}

	// remove alias if present
	var aliases []config.AliasConfig
	for _, alias := range conf.Aliases {
		if alias.Name != c.Args().First() {
			aliases = append(aliases, alias)
		}
	}

	aliases = append(conf.Aliases, config.AliasConfig{
		Name:        c.Args().First(),
		Synchronize: status,
	})
	conf.Aliases = aliases

	if err := odc.saveConfig(conf); err != nil {
		odc.logger.Err(err).Msg("error while saving config.")
		return err
	}

	m := odc.logger.Info().Str("Domain", c.Args().First())
	if status {
		m.Msg("enable synchronization.")
	} else {
		m.Msg("disable synchronization.")
	}

	return nil
}

func (odc *OpenDYDNSCLI) synchronize(_ *cli.Context) error {
	token, err := odc.getToken()
	if err != nil {
		odc.logger.Err(err).Msg("error while getting JWT token.")
		return err
	}

	ip, err := odc.getRemoteIP()
	if err != nil {
		odc.logger.Err(err).Msg("error while getting remote IP.")
		return err
	}

	apiClient := client.NewClient(odc.conf.APIAddr)

	for _, alias := range odc.conf.Aliases {
		if _, err := apiClient.UpdateAlias(token, proto.AliasDto{
			Domain: alias.Name,
			Value:  ip,
		}); err != nil {
			odc.logger.Err(err).Str("Domain", alias.Name).Str("Value", ip).Msg("error while updating alias.")
		} else {
			odc.logger.Info().Str("Domain", alias.Name).Str("Value", ip).Msg("successfully updated alias.")
		}
	}

	return nil
}

func (odc *OpenDYDNSCLI) saveConfig(conf config.Config) error {
	odc.conf = conf
	return config.Save(odc.conf, odc.confPath)
}

func (odc *OpenDYDNSCLI) getToken() (proto.TokenDto, error) {
	if odc.conf.Token == "" {
		return proto.TokenDto{}, fmt.Errorf("not logged in")
	}

	return proto.TokenDto{Token: odc.conf.Token}, nil
}

func (odc *OpenDYDNSCLI) getRemoteIP() (string, error) {
	c := resty.New()
	r, err := c.R().Get("https://ifconfig.me/ip")
	if err != nil {
		return "", err
	}

	return r.String(), nil
}
