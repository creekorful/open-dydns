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
	return &cli.App{
		Name:    "opendydns-cli",
		Usage:   "The OpenDyDNS CLI",
		Authors: []*cli.Author{{Name: "Aloïs Micard", Email: "alois@micard.lu"}},
		Version: "0.1.0",
		Before:  odc.before,
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
				ArgsUsage: "<EMAIL>",
				Usage:     "Authenticate against an OpenDyDNS daemon",
				Action:    odc.login,
			},
			{
				Name:   "ls",
				Usage:  "List current DyDNS aliases",
				Action: odc.ls,
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
		},
	}
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
			return err
		}

		return nil
	}

	// Load the configuration file
	conf, err := config.Load(configFile)
	if err != nil {
		return err
	}

	// Store configuration file
	odc.conf = conf
	odc.confPath = configFile

	return nil
}

func (odc *OpenDYDNSCLI) login(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing EMAIL")
	}

	// check if not already logged in
	if odc.conf.Token != "" {
		return fmt.Errorf("already logged in")
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
		return err
	}

	// Save token in config file
	odc.conf.Token = token.Token
	if err := odc.saveConfig(odc.conf); err != nil {
		return err
	}

	odc.logger.Info().Str("Email", c.Args().First()).Msg("successfully authenticated.")

	return nil // TODO implement
}

func (odc *OpenDYDNSCLI) ls(_ *cli.Context) error {
	token, err := odc.getToken()
	if err != nil {
		return err
	}

	apiClient := client.NewClient(odc.conf.APIAddr)

	aliases, err := apiClient.GetAliases(token)
	if err != nil {
		return err
	}

	if len(aliases) == 0 {
		fmt.Println("no aliases found")
		return nil
	}

	// TODO use proper table
	for _, alias := range aliases {
		fmt.Printf("%s -> %s\n", alias.Domain, alias.Value)
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
		return err
	}

	ip, err := odc.getRemoteIP()
	if err != nil {
		return err
	}

	alias, err := apiClient.RegisterAlias(token, proto.AliasDto{
		Domain: name,
		Value:  ip,
	})

	if err != nil {
		return err
	}

	odc.logger.Info().Str("Alias", alias.Domain).Msg("successfully created alias.")
	return nil
}

func (odc *OpenDYDNSCLI) rm(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing ALIAS")
	}

	name := c.Args().First()

	apiClient := client.NewClient(odc.conf.APIAddr)

	token, err := odc.getToken()
	if err != nil {
		return err
	}

	if err := apiClient.DeleteAlias(token, name); err != nil {
		return err
	}

	odc.logger.Info().Str("Alias", name).Msg("successfully deleted alias.")
	return nil
}

func (odc *OpenDYDNSCLI) setIP(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return fmt.Errorf("missing ALIAS IP")
	}

	alias := c.Args().First()
	ip := c.Args().Get(1)

	apiClient := client.NewClient(odc.conf.APIAddr)

	token, err := odc.getToken()
	if err != nil {
		return err
	}

	al, err := apiClient.UpdateAlias(token, proto.AliasDto{
		Domain: alias,
		Value:  ip,
	})

	if err != nil {
		return err
	}

	odc.logger.Info().Str("Alias", al.Value).Str("Value", al.Value).Msg("successfully deleted alias.")
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
