package daemon

import (
	"errors"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/database"
	"github.com/creekorful/open-dydns/internal/proto"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
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
	UpdateAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error)
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
	log.Info().Str("Driver", c.DatabaseConfig.Driver).Msg("database connection established!")

	d := &daemon{
		conn: conn,
	}

	// TODO remove below code
	if _, err := conn.FindUser("lunamicard@gmail.com"); errors.As(err, &gorm.ErrRecordNotFound) {
		pass, err := d.hashPassword("test")
		if err != nil {
			return nil, err
		}

		if _, err := conn.CreateUser("lunamicard@gmail.com", pass); err != nil {
			return nil, err
		}
	}

	return d, nil
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
	}, nil
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

func (d *daemon) UpdateAlias(userCtx proto.UserContext, alias proto.AliasDto) (proto.AliasDto, error) {
	al, err := d.findUserAlias(alias.Domain, userCtx.UserID)
	if err != nil {
		return proto.AliasDto{}, err
	}

	// Update the alias
	al.Value = alias.Value
	al, err = d.conn.UpdateAlias(al)
	if err != nil {
		log.Err(err).Msg("error while updating alias.")
		return proto.AliasDto{}, err
	}

	return newAliasDto(al), err
}

func (d *daemon) DeleteAlias(userCtx proto.UserContext, aliasName string) error {
	if err := d.conn.DeleteAlias(aliasName, userCtx.UserID); err != nil {
		log.Warn().Str("Alias", aliasName).Uint("UserID", userCtx.UserID).Msg("unable to delete alias.")
		return err
	}
	// TODO trigger linked code

	log.Debug().Str("Alias", aliasName).Uint("UserID", userCtx.UserID).Msg("successfully deleted alias.")

	return nil
}

func (d *daemon) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
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

func (d *daemon) findUserAlias(name string, userID uint) (database.Alias, error) {
	al, err := d.conn.FindAlias(name)
	if err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			return database.Alias{}, ErrAliasNotFound
		} else {
			return database.Alias{}, err
		}
	}

	if al.UserID != userID {
		return database.Alias{}, ErrAliasNotFound
	}

	return al, nil
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
