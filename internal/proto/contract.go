package proto

type APIContract interface {
	// POST /sessions
	Authenticate(cred CredentialsDto) (TokenDto, error)
	// GET /aliases
	GetAliases(token TokenDto) ([]AliasDto, error)
	// POST /aliases
	RegisterAlias(token TokenDto, alias AliasDto) (AliasDto, error)
	// PUT /aliases/{name}
	UpdateAlias(token TokenDto, alias AliasDto) (AliasDto, error)
	// DELETE /aliases/{name}
	DeleteAlias(token TokenDto, name string) error
}

type AliasDto struct {
	Domain string `json:"domain"`
	Value  string `json:"value"`
}

type CredentialsDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenDto struct {
	Token string `json:"token"`
}

// TODO make my own error mapper
type ErrorDto struct {
	Message string `json:"message"`
}

func (e ErrorDto) Error() string {
	return e.Message
}

type UserContext struct {
	UserID uint `json:"UserID"`
}
