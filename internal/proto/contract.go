package proto

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
