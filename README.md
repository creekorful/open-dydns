# open-dydns

Open DyDNS is a free & open source DyDNS service.
It allows users to create their DyDNS services with their own domains and provide access
to them trough a secure, authenticated API.

Open DyDNS is built upon two components:

## opendydnsd

OpenDyDNSD is the daemon service running on your webserver, it will use a special dedicated config file
to read the supported / configured domains & dial with a database to manage access control.

The daemon is exposed using an authenticated REST API with JWT authentication.
The daemon configuration is only configurable by editing the config file, not trough the API.

### API contract

#### POST /sessions

Authenticate against the API.
This will either return the token with an HTTP 200, 
or an error response if invalid credentials are supplied.

Request Body:

```json
{
  "username": "test",
  "password": "test"
}
```

Response Body:

```json
{
  "token": "token"
}
```

#### GET /aliases

Get the aliases of logged user.

Response Body:

```json
[
  {
    "domain": "foo.example.org",
    "value": "127.0.0.1"
  }
]
```

#### POST /aliases

Register given aliases for logged user.
It either returns the created resource with an HTTP 201,
or an error response if something happened.

Request Body:

```json
{
  "domain": "bar.example.org",
  "value": "127.0.0.1"
}
```

Response Body:

```json
{
  "domain": "bar.example.org",
  "value": "127.0.0.1"
}
```

#### DELETE /aliases/{alias}

Delete given alias for logged user.
It either returns an HTTP 204 is operation is successful,
or an error response if something happened.

### The configuration file

Below is a example of the configuration file:

```yaml
```

## opendydns-cli

OpenDyDNS-CLI is a CLI used to dial with the daemon. It uses the REST API.

Each time the CLI is installed on a computer, a new access token must be registered using the login command.

### Commands

```
$ opendydns-cli login <username>
```

This command will prompt for the user password and then tries to authenticate it and save the auth token
on the system.

```
$ opendydns-cli ls
```

This command will list the user current DyDNS aliases with linked current IP.
This will also enable the alias for given computer and synchronize the IP.

```
$ opendydns-cli add <alias>
```

This command will register given alias if possible and associated with current computer.

```
$ opendydns-cli rm <alias>
```

This command will delete given alias (marking it available for others).

```
$ opendydns-cli enable <alias>
```

Enable given alias on this computer. (I.e the computer will synchronize it using its IP address)
By default all aliases are disabled, to prevent any service disruption when adding a new computer.

```
$ opendydns-cli disable <alias>
```

Disable given alias for this computer. This will stop IP synchronization.

```
$ opendydns-cli set-ip <alias> <ip>
```

Override the IP value for given alias. This works with both IPv4 and Ipv6.

```
$ opendydns-cli synchronize
```

This command will synchronize the current IP with linked / active aliases.
This is generally run by a Cron job.