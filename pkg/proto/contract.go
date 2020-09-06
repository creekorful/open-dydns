package proto

import "github.com/labstack/echo/v4"

//go:generate mockgen -source contract.go -destination=./contract_mock.go -package=proto

// ErrAliasTaken is returned when the wanted alias is already taken by someone else
var ErrAliasTaken = echo.NewHTTPError(409, "alias already taken")

// ErrAliasAlreadyExist is returned when user already own the wanted alias
var ErrAliasAlreadyExist = echo.NewHTTPError(409, "alias already exist")

// ErrAliasNotFound is returned when the wanted alias cannot be found
var ErrAliasNotFound = echo.NewHTTPError(404, "alias not found")

// ErrInvalidParameters is returned when the given request is invalid
var ErrInvalidParameters = echo.NewHTTPError(400, "invalid request parameter(s)")

// ErrDomainNotFound is returned when the alias to register use non supported / not existing domain
var ErrDomainNotFound = echo.NewHTTPError(404, "requested domain not found")

// APIContract defined the API served by the Daemon
type APIContract interface {
	// Authenticate user using given credential
	// this either return the JWT token or an error if something goes wrong
	// POST /sessions
	Authenticate(cred CredentialsDto) (TokenDto, error)
	// GetAliases return user current aliases
	// GET /aliases
	GetAliases(token TokenDto) ([]AliasDto, error)
	// RegisterAlias register a new alias for the user
	// POST /aliases
	RegisterAlias(token TokenDto, alias AliasDto) (AliasDto, error)
	// UpdateAlias update the user existing alias
	// PUT /aliases/{name}
	UpdateAlias(token TokenDto, alias AliasDto) (AliasDto, error)
	// DeleteAlias delete the user given alias
	// DELETE /aliases/{name}
	DeleteAlias(token TokenDto, name string) error

	// GetDomains return the list of available / supported domains
	// for alias creation
	// GET /domains
	GetDomains(token TokenDto) ([]DomainDto, error)
}

// AliasDto represent a DyDNS alias
type AliasDto struct {
	Domain string `json:"domain"`
	Value  string `json:"value"`
}

// CredentialsDto represent the credentials
// when issuing a authentication request
type CredentialsDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenDto represent the object that encapsulate the JWT token
// when issuing a authentication request
type TokenDto struct {
	Token string `json:"token"`
}

// DomainDto represent a domain usable to create alias
// on the Daemon
type DomainDto struct {
	Domain string `json:"domain"`
}

// ErrorDto is the generic error response in case of API error
// TODO make my own error mapper
type ErrorDto struct {
	Message string `json:"message"`
}

func (e ErrorDto) Error() string {
	return e.Message
}

// UserContext represent the JWT token payload
// and identify the logged in user in secured endpoints
type UserContext struct {
	UserID uint `json:"UserID"`
}
