package cli

import (
	"fmt"
	"github.com/creekorful/open-dydns/internal/opendydnscli/client"
	"github.com/creekorful/open-dydns/internal/opendydnscli/config"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/rs/zerolog"
)

// ErrBadRequest is returned when function is calling with missing parameters
var ErrBadRequest = fmt.Errorf("missing parameters")

// ErrAlreadyLoggedIn is returned when trying to log-in but already logged in
var ErrAlreadyLoggedIn = fmt.Errorf("already logged in")

// AliasStatus represent an alias as viewed by the CLI app
type AliasStatus struct {
	proto.AliasDto
	Synchronize bool
}

// CLI represent a instance of the cli application
type CLI interface {
	Authenticate(cred proto.CredentialsDto) (proto.TokenDto, error)
	GetAliases() ([]AliasStatus, error)
	RegisterAlias(alias proto.AliasDto) (proto.AliasDto, error)
	UpdateAlias(alias proto.AliasDto) (proto.AliasDto, error)
	DeleteAlias(aliasName string) error
	GetDomains() ([]proto.DomainDto, error)
	SetSynchronize(aliasName string, status bool) error
	Synchronize(IP string) error
}

type cli struct {
	tok          proto.TokenDto
	logger       *zerolog.Logger
	conf         config.Config
	confProvider config.Provider
	apiClient    proto.APIContract
}

// NewCLI instantiate a new CLI instance
func NewCLI(confPath string, logger *zerolog.Logger) (CLI, error) {
	provider := config.NewFileProvider(confPath)

	// Load the configuration file
	conf, err := provider.Load()
	if err != nil {
		return nil, err
	}

	if !conf.Valid() {
		return nil, fmt.Errorf("invalid config file")
	}

	return &cli{
		tok:          proto.TokenDto{Token: conf.Token},
		logger:       logger,
		conf:         conf,
		confProvider: provider,
		apiClient:    client.NewClient(conf.APIAddr),
	}, nil
}

func (c *cli) Authenticate(cred proto.CredentialsDto) (proto.TokenDto, error) {
	if cred.Email == "" || cred.Password == "" {
		return proto.TokenDto{}, ErrBadRequest
	}

	// check if not already logged in
	if c.conf.Token != "" {
		return proto.TokenDto{}, ErrAlreadyLoggedIn
	}

	token, err := c.apiClient.Authenticate(cred)
	if err != nil {
		return proto.TokenDto{}, err
	}

	// save token
	c.conf.Token = token.Token
	if err := c.saveConfig(); err != nil {
		return proto.TokenDto{}, err
	}

	return proto.TokenDto{Token: c.conf.Token}, nil
}

func (c *cli) GetAliases() ([]AliasStatus, error) {
	aliases, err := c.apiClient.GetAliases(c.tok)
	if err != nil {
		return nil, err
	}

	var aliasStatuses []AliasStatus

	for _, alias := range aliases {
		status := false
		for name, conf := range c.conf.Aliases {
			if name == alias.Domain {
				status = conf.Synchronize
			}
		}
		aliasStatuses = append(aliasStatuses, AliasStatus{
			AliasDto:    alias,
			Synchronize: status,
		})
	}

	return aliasStatuses, nil
}

func (c *cli) RegisterAlias(alias proto.AliasDto) (proto.AliasDto, error) {
	if alias.Domain == "" || alias.Value == "" {
		return proto.AliasDto{}, ErrBadRequest
	}

	return c.apiClient.RegisterAlias(c.tok, alias)
}

func (c *cli) UpdateAlias(alias proto.AliasDto) (proto.AliasDto, error) {
	if alias.Domain == "" || alias.Value == "" {
		return proto.AliasDto{}, ErrBadRequest
	}

	return c.apiClient.UpdateAlias(c.tok, alias)
}

func (c *cli) DeleteAlias(aliasName string) error {
	if aliasName == "" {
		return ErrBadRequest
	}

	return c.apiClient.DeleteAlias(c.tok, aliasName)
}

func (c *cli) GetDomains() ([]proto.DomainDto, error) {
	return c.apiClient.GetDomains(c.tok)
}

func (c *cli) SetSynchronize(aliasName string, status bool) error {
	conf := c.conf
	if conf.Aliases == nil {
		conf.Aliases = map[string]config.AliasConfig{}
	}

	var aliasConfig config.AliasConfig
	if v, exist := conf.Aliases[aliasName]; exist {
		aliasConfig = v
	}

	aliasConfig.Synchronize = status
	conf.Aliases[aliasName] = aliasConfig

	if err := c.saveConfig(); err != nil {
		return err
	}

	return nil
}

func (c *cli) Synchronize(ip string) error {
	for name, conf := range c.conf.Aliases {
		if !conf.Synchronize {
			continue
		}

		if _, err := c.UpdateAlias(proto.AliasDto{
			Domain: name,
			Value:  ip,
		}); err != nil {
			c.logger.Err(err).Str("Domain", name).Str("Value", ip).Msg("error while updating alias.")
		} else {
			c.logger.Info().Str("Domain", name).Str("Value", ip).Msg("successfully updated alias.")
		}
	}

	return nil
}

func (c *cli) saveConfig() error {
	return c.confProvider.Save(c.conf)
}
