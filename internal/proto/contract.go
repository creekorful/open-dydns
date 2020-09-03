package proto

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

type UserContext struct {
	UserID uint `json:"userId"`
}
