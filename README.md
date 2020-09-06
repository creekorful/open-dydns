# open-dydns

[![Go Report Card](https://goreportcard.com/badge/github.com/creekorful/open-dydns)](https://goreportcard.com/report/github.com/creekorful/open-dydns)
[![codecov](https://codecov.io/gh/creekorful/open-dydns/branch/master/graph/badge.svg)](https://codecov.io/gh/creekorful/open-dydns)

Open DyDNS is a free & open source DyDNS service.
It allows users to create their own DyDNS services with custom domains and provide access
to them trough a secure, authenticated API.

Open DyDNS is built upon two components:

## opendydnsd

OpenDyDNSD is the daemon service running on your sever, it will use a special dedicated config file
to read the supported / configured domains & dial with a database to manage access control.

The daemon is exposed using an authenticated REST API with JWT authentication.
The daemon configuration is only configurable by editing the config file, not trough the API.

### API contract

Here's the Go definition of the API contract.

```go
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
	// GET /domains
	GetDomains(token TokenDto) ([]DomainDto, error)
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

type ErrorDto struct {
	Message string `json:"message"`
}

type DomainDto struct {
	Domain string `json:"domain"`
}
```

### The configuration file

Below is an example of the configuration file:

```toml
[ApiConfig]
  ListenAddr = "127.0.0.1:8888"
  SigningKey = "TEST"

[DaemonConfig]

[DatabaseConfig]
  DSN = "test.db"
  Driver = "sqlite"
```

## opendydns-cli

OpenDyDNS-CLI is a CLI used to dial with the daemon. It uses the REST API.

Each time the CLI is installed on a computer, a new access token must be registered using the login command.

### Commands

```
$ opendydns-cli login <email>
```

This command will prompt for the user password and then tries to authenticate it and save the JWT token
on the system.

```
$ opendydns-cli ls <what>
```

This command will list the available resources.
Possible resources: domain or alias. Default is alias.

```
$ opendydns-cli add <alias>
```

This command will register given alias if possible and associated with current computer.
This will also enable the alias for given computer and synchronize the IP.

```
$ opendydns-cli rm <alias>
```

This command will delete given alias (marking it available for others).

```
$ opendydns-cli set-synchronize <alias> <true/false>
```

Enable IP synchronization for this alias.
Please note that by default synchronization is disable, to prevent any service disruption when adding a new computer.

```
$ opendydns-cli set-ip <alias> <ip>
```

Override the IP value for given alias. This works with both IPv4 and Ipv6.

```
$ opendydns-cli sync
```

This command will synchronize the current IP with linked / active aliases.
This is generally run by a Cron job.