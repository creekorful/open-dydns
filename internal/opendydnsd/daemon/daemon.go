package daemon

import (
	"errors"
	"fmt"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/database"
	"github.com/creekorful/open-dydns/internal/opendydnsd/dns"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
)

//go:generate mockgen -source daemon.go -destination=../daemon_mock/daemon_mock.go -package=daemon_mock

// Daemon represent OpenDyDNSD
type Daemon interface {
	CreateUser(cred proto.CredentialsDto) (proto.UserContext, error)
	Authenticate(cred proto.CredentialsDto) (proto.UserContext, error)
	GetAliases(userCtx proto.UserContext) ([]proto.AliasDto, error)
	RegisterAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error)
	UpdateAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error)
	DeleteAlias(userCtx proto.UserContext, aliasName string) error
	GetDomains(userCtx proto.UserContext) ([]proto.DomainDto, error)
	Logger() *zerolog.Logger
}

type daemon struct {
	conn        database.Connection
	logger      *zerolog.Logger
	config      config.DaemonConfig
	dnsProvider dns.Provider
}

// NewDaemon return a new Daemon instance with given configuration
func NewDaemon(c config.Config, logger *zerolog.Logger) (Daemon, error) {
	logger.Debug().Msg("connecting to the database.")
	conn, err := database.OpenConnection(c.DatabaseConfig, logger)
	if err != nil {
		return nil, err
	}
	logger.Info().Str("Driver", c.DatabaseConfig.Driver).Msg("database connection established!")

	d := &daemon{
		conn:        conn,
		logger:      logger,
		config:      c.DaemonConfig,
		dnsProvider: dns.NewProvider(),
	}

	return d, nil
}

func (d *daemon) CreateUser(cred proto.CredentialsDto) (proto.UserContext, error) {
	if cred.Email == "" || cred.Password == "" {
		d.logger.Warn().Msg("invalid create user request: bad request.")
		return proto.UserContext{}, proto.ErrInvalidParameters
	}

	// Make sure user doesn't already exist
	_, err := d.conn.FindUser(cred.Email)
	if err != nil && !errors.As(err, &gorm.ErrRecordNotFound) {
		d.logger.Err(err).Msg("error while fetching database.")
		return proto.UserContext{}, err
	} else if err == nil {
		d.logger.Warn().Msg("email address already taken.")
		return proto.UserContext{}, proto.ErrInvalidParameters // not 409 to prevent email discovery
	}

	// Doesn't exist yet!
	pass, err := d.hashPassword(cred.Password)
	if err != nil {
		return proto.UserContext{}, err
	}

	if _, err := d.conn.CreateUser(cred.Email, pass); err != nil {
		return proto.UserContext{}, err
	}

	return d.Authenticate(cred)
}

func (d *daemon) Authenticate(cred proto.CredentialsDto) (proto.UserContext, error) {
	if cred.Email == "" || cred.Password == "" {
		d.logger.Warn().Msg("invalid authentication request: bad request.")
		return proto.UserContext{}, proto.ErrInvalidParameters
	}

	user, err := d.conn.FindUser(cred.Email)
	if errors.As(err, &gorm.ErrRecordNotFound) {
		return proto.UserContext{}, proto.ErrInvalidParameters // not 404 to prevent email discovery
	}
	if err != nil {
		return proto.UserContext{}, err
	}

	// Validate the password
	if !d.validatePassword(user.Password, cred.Password) {
		d.logger.Warn().Msg("invalid authentication request: invalid password.")
		return proto.UserContext{}, proto.ErrInvalidParameters // not 404 to prevent email discovery
	}

	d.logger.Debug().Str("Email", user.Email).Msg("successfully authenticated.")

	return proto.UserContext{
		UserID: user.ID,
	}, nil
}

func (d *daemon) GetAliases(userCtx proto.UserContext) ([]proto.AliasDto, error) {
	aliases, err := d.conn.FindUserAliases(userCtx.UserID)

	if err != nil && !errors.As(err, &gorm.ErrRecordNotFound) {
		d.logger.Err(err).Msg("error while fetching database.")
		return nil, err
	}

	var aliasesDto []proto.AliasDto
	for _, alias := range aliases {
		aliasesDto = append(aliasesDto, newAliasDto(alias))
	}

	return aliasesDto, nil
}

func (d *daemon) RegisterAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error) {
	if !isAliasValid(alias) {
		d.logger.Warn().Msg("invalid register alias request: bad request.")
		return proto.AliasDto{}, proto.ErrInvalidParameters
	}

	a := newAlias(alias)

	provisioner, domainConf, err := d.findDNSProvisioner(a.Domain)
	if err != nil {
		d.logger.Err(err).Str("Domain", a.Domain).Msg("domain is not supported.")
		return proto.AliasDto{}, proto.ErrDomainNotFound
	}

	res, err := d.conn.FindAlias(a.Host, a.Domain)

	// technical error
	if err != nil && !errors.As(err, &gorm.ErrRecordNotFound) {
		d.logger.Err(err).Msg("error while fetching database.")
		return proto.AliasDto{}, err
	}

	// record already exist
	if err == nil {
		if res.UserID != userCtx.UserID {
			d.logger.Debug().Msg("alias taken.")
			return proto.AliasDto{}, proto.ErrAliasTaken
		}

		d.logger.Debug().Msg("alias already exist.")
		return proto.AliasDto{}, proto.ErrAliasAlreadyExist
	}

	// alias available: perform registration
	host, domain := getRealHostAndDomain(alias, domainConf)
	if err := provisioner.AddRecord(host, domain, a.Value); err != nil {
		d.logger.Err(err).
			Str("Domain", domain).
			Str("Host", host).
			Str("Value", a.Value).
			Msg("error while adding DNS record.")
		return proto.AliasDto{}, err
	}

	a, err = d.conn.CreateAlias(newAlias(alias), userCtx.UserID)
	if err != nil {
		return proto.AliasDto{}, err
	}
	d.logger.Debug().
		Uint("UserID", userCtx.UserID).
		Str("Domain", a.Domain).
		Str("Host", a.Host).
		Str("Value", a.Value).
		Msg("new alias created.")

	return newAliasDto(a), nil
}

func (d *daemon) UpdateAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error) {
	if !isAliasValid(alias) {
		d.logger.Warn().Msg("invalid update alias request: bad request.")
		return proto.AliasDto{}, proto.ErrInvalidParameters
	}

	al, err := d.findUserAlias(alias, userCtx.UserID)
	if err != nil {
		return proto.AliasDto{}, err
	}

	// Update the alias
	updateAlias(&al, alias)

	provisioner, domainConf, err := d.findDNSProvisioner(al.Domain)
	if err != nil {
		d.logger.Err(err).Msg("error while finding DNS provisioner.")
		return proto.AliasDto{}, err
	}

	host, domain := getRealHostAndDomain(alias, domainConf)
	if err := provisioner.UpdateRecord(host, domain, al.Value); err != nil {
		d.logger.Err(err).
			Str("Domain", domain).
			Str("Host", host).
			Str("Value", al.Value).
			Msg("error while updating DNS record.")
		return proto.AliasDto{}, err
	}

	al, err = d.conn.UpdateAlias(al)
	if err != nil {
		d.logger.Err(err).Msg("error while updating alias.")
		return proto.AliasDto{}, err
	}

	d.logger.Debug().
		Uint("UserID", userCtx.UserID).
		Str("Domain", al.Domain).
		Str("Host", al.Host).
		Str("Value", alias.Value).
		Msg("successfully updated alias.")

	return newAliasDto(al), err
}

func (d *daemon) DeleteAlias(userCtx proto.UserContext, aliasName string) error {
	a := newAlias(proto.AliasDto{Domain: aliasName})

	provisioner, domainConf, err := d.findDNSProvisioner(a.Domain)
	if err != nil {
		d.logger.Err(err).Msg("error while finding DNS provisioner.")
		return err
	}

	host, domain := getRealHostAndDomain(proto.AliasDto{Domain: aliasName}, domainConf)
	if err := provisioner.DeleteRecord(host, domain); err != nil {
		d.logger.Err(err).
			Str("Domain", domain).
			Str("Host", host).
			Msg("error while deleting DNS record.")
		return err
	}

	if err := d.conn.DeleteAlias(a.Host, a.Domain, userCtx.UserID); err != nil {
		d.logger.Warn().
			Str("Domain", a.Domain).
			Str("Host", a.Host).
			Msg("unable to delete alias.")
		return err
	}

	d.logger.Debug().
		Uint("UserID", userCtx.UserID).
		Str("Domain", a.Domain).
		Str("Host", a.Host).
		Msg("successfully deleted alias.")

	return nil
}

func (d *daemon) GetDomains(_ proto.UserContext) ([]proto.DomainDto, error) {
	var domains []proto.DomainDto

	for _, dnsProvisioner := range d.config.DNSProvisioners {
		for _, domain := range dnsProvisioner.Domains {
			domains = append(domains, proto.DomainDto{
				Domain: domain.String(),
			})
		}
	}

	return domains, nil
}

func (d *daemon) Logger() *zerolog.Logger {
	return d.logger
}

func (d *daemon) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		d.logger.Err(err).Msg("error while hashing password.")
		return "", err
	}

	return string(hash), nil
}

func (d *daemon) validatePassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		return false
	}

	return true
}

func (d *daemon) findUserAlias(alias proto.AliasDto, userID uint) (database.Alias, error) {
	a := newAlias(alias)
	al, err := d.conn.FindAlias(a.Host, a.Domain)
	if err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			return database.Alias{}, proto.ErrAliasNotFound
		}

		return database.Alias{}, err
	}

	if al.UserID != userID {
		return database.Alias{}, proto.ErrAliasNotFound
	}

	return al, nil
}

func (d *daemon) findDNSProvisioner(domain string) (dns.Provisioner, config.DomainConfig, error) {
	for _, dnsProvisioner := range d.config.DNSProvisioners {
		for _, domainConf := range dnsProvisioner.Domains {
			if domainConf.String() == domain {
				p, err := d.dnsProvider.GetProvisioner(dnsProvisioner.Name, dnsProvisioner.Config)
				return p, domainConf, err
			}
		}
	}

	return nil, config.DomainConfig{}, fmt.Errorf("no DNS provisioner found for domain %s", domain)
}

// Alias -> AliasDto
func newAliasDto(alias database.Alias) proto.AliasDto {
	return proto.AliasDto{
		Domain: fmt.Sprintf("%s.%s", alias.Host, alias.Domain),
		Value:  alias.Value,
	}
}

// AliasDto -> Alias
func newAlias(alias proto.AliasDto) database.Alias {
	parts := strings.Split(alias.Domain, ".")
	return database.Alias{
		Host:   parts[0],
		Domain: strings.Replace(alias.Domain, parts[0]+".", "", 1),
		Value:  alias.Value,
	}
}

// Update an existing alias using given DTO
func updateAlias(alias *database.Alias, dto proto.AliasDto) {
	a := newAlias(dto)

	alias.Host = a.Host
	alias.Value = a.Value
}

func isAliasValid(alias proto.AliasDto) bool {
	// TODO make sure value is valid IPv4 / IpV6
	return alias.Domain != "" && strings.Count(alias.Domain, ".") >= 2 && alias.Value != ""
}

func getRealHostAndDomain(alias proto.AliasDto, domainConf config.DomainConfig) (string, string) {
	host := strings.Replace(alias.Domain, "."+domainConf.Domain, "", 1)
	return host, domainConf.Domain
}
