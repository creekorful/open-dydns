package daemon

import (
	"errors"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/database"
	"github.com/creekorful/open-dydns/internal/proto"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var ErrUserNotFound = echo.NewHTTPError(404, "user not found")
var ErrAliasTaken = echo.NewHTTPError(409, "alias already taken")
var ErrAliasAlreadyExist = echo.NewHTTPError(409, "alias already exist")
var ErrAliasNotFound = echo.NewHTTPError(404, "alias not found")
var ErrInvalidParameters = echo.NewHTTPError(404, "invalid request parameter(s)")

type Daemon interface {
	Authenticate(cred proto.CredentialsDto) (proto.UserContext, error)
	GetAliases(userCtx proto.UserContext) ([]proto.AliasDto, error)
	RegisterAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error)
	DeleteAlias(userCtx proto.UserContext, aliasName string) error
}

type daemon struct {
	conn database.Connection
}

func NewDaemon(c config.Config) (Daemon, error) {
	log.Debug().Msg("connecting to the database.")
	conn, err := database.OpenConnection(c.DatabaseConfig)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("database connection established!")

	return &daemon{
		conn: conn,
	}, nil
}

func (d *daemon) Authenticate(cred proto.CredentialsDto) (proto.UserContext, error) {
	if cred.Email == "" || cred.Password == "" {
		log.Warn().Msg("invalid authentication request: bad request.")
		return proto.UserContext{}, ErrInvalidParameters
	}

	user, err := d.conn.FindUser(cred.Email)
	if errors.As(err, &gorm.ErrRecordNotFound) {
		return proto.UserContext{}, ErrUserNotFound
	}
	if err != nil {
		return proto.UserContext{}, err
	}

	// Validate the password
	if !d.validatePassword(user.Password, cred.Password) {
		log.Warn().Msg("invalid authentication request: invalid password.")
		return proto.UserContext{}, ErrUserNotFound
	}

	log.Debug().Str("Email", user.Email).Msg("successfully authenticated.")

	return proto.UserContext{
		UserID: user.ID,
	}, nil // TODO
}

func (d *daemon) GetAliases(userCtx proto.UserContext) ([]proto.AliasDto, error) {
	aliases, err := d.conn.FindUserAliases(userCtx.UserID)

	if err != nil && !errors.As(err, &gorm.ErrRecordNotFound) {
		log.Err(err).Msg("error while fetching database.")
		return nil, err
	}

	var aliasesDto []proto.AliasDto
	for _, alias := range aliases {
		aliasesDto = append(aliasesDto, newAliasDto(alias))
	}

	return aliasesDto, nil
}

func (d *daemon) RegisterAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error) {
	if alias.Domain == "" || alias.Value == "" {
		log.Warn().Msg("invalid register alias request: bad request.")
		return proto.AliasDto{}, ErrInvalidParameters
	}

	a, err := d.conn.FindAlias(alias.Domain)

	// technical error
	if err != nil && !errors.As(err, &gorm.ErrRecordNotFound) {
		log.Err(err).Msg("error while fetching database.")
		return proto.AliasDto{}, err
	}

	// record already exist
	if err == nil {
		if a.UserID != userCtx.UserID {
			log.Debug().Msg("alias taken.")
			return proto.AliasDto{}, ErrAliasTaken
		} else {
			log.Debug().Msg("alias already exist.")
			return proto.AliasDto{}, ErrAliasAlreadyExist
		}
	}

	// alias available
	// TODO trigger provisioning linked code

	a, err = d.conn.CreateAlias(newAlias(alias), userCtx.UserID)
	if err != nil {
		return proto.AliasDto{}, err
	}
	log.Debug().Str("Domain", a.Domain).Msg("new alias created.")

	return newAliasDto(a), nil
}

func (d *daemon) DeleteAlias(userCtx proto.UserContext, aliasName string) error {
	// TODO trigger linked code
	return d.conn.DeleteAlias(aliasName, userCtx.UserID)
}

func (d *daemon) encryptPassword(password string) (string, error) {
	return password, nil // TODO implement
}

func (d *daemon) validatePassword(encryptedPassword, plainPassword string) bool {
	return encryptedPassword == plainPassword // TODO implement
}

// Alias -> AliasDto
func newAliasDto(alias database.Alias) proto.AliasDto {
	return proto.AliasDto{
		Domain: alias.Domain,
		Value:  alias.Value,
	}
}

// AliasDto -> Alias
func newAlias(alias proto.AliasDto) database.Alias {
	return database.Alias{
		Model:  gorm.Model{},
		Domain: alias.Domain,
		Value:  alias.Value,
	}
}
