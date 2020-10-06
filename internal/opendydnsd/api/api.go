package api

import (
	"context"
	"fmt"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/daemon"
	"github.com/creekorful/open-dydns/proto"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"net/http"
	"strings"
)

// API represent the Daemon REST API
type API struct {
	e      *echo.Echo
	conf   config.APIConfig
	logger *zerolog.Logger
}

// NewAPI return a new API instance, wrapped around given Daemon instance
// and with given config
func NewAPI(d daemon.Daemon, conf config.APIConfig) (*API, error) {
	// Configure echo
	e := echo.New()
	e.Logger.SetOutput(ioutil.Discard)

	// Determinate if should run HTTPS
	if conf.SSLEnabled() {
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(conf.Hostname)
		e.AutoTLSManager.Cache = autocert.DirCache(conf.CertCacheDir)
	}

	// Create the API
	a := API{
		e:      e,
		conf:   conf,
		logger: d.Logger(),
	}

	// Register global middlewares
	e.Use(newZeroLogMiddleware(d.Logger()))

	// Register per-route middlewares
	authMiddleware := getAuthMiddleware(a.conf.SigningKey)

	// Register endpoints
	e.POST("/sessions", a.authenticate(d))
	e.GET("/aliases", a.getAliases(d), authMiddleware)
	e.POST("/aliases", a.registerAlias(d), authMiddleware)
	e.PUT("/aliases", a.updateAlias(d), authMiddleware)
	e.DELETE("/aliases/:name", a.deleteAlias(d), authMiddleware)
	e.GET("/domains", a.getDomains(d), authMiddleware)

	return &a, nil
}

func (a *API) authenticate(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cred proto.CredentialsDto
		if err := c.Bind(&cred); err != nil {
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		userCtx, err := d.Authenticate(cred)
		if err != nil {
			return err
		}

		// Create the JWT token
		token, err := makeToken(userCtx, a.conf.SigningKey, a.conf.TokenTTL)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, token)
	}
}

func (a *API) getAliases(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		aliases, err := d.GetAliases(userCtx)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, aliases)
	}
}

func (a *API) registerAlias(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		var alias proto.AliasDto
		if err := c.Bind(&alias); err != nil {
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		alias, err := d.RegisterAlias(userCtx, alias)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusCreated, alias)
	}
}

func (a *API) updateAlias(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		var alias proto.AliasDto
		if err := c.Bind(&alias); err != nil {
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		alias, err := d.UpdateAlias(userCtx, alias)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, alias)
	}
}

func (a *API) deleteAlias(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		alias := c.Param("name")

		if err := d.DeleteAlias(userCtx, alias); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

func (a *API) getDomains(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		domains, err := d.GetDomains(userCtx)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, domains)
	}
}

// Start the API server
func (a *API) Start(address string) error {
	// determinate if should run HTTPS
	if a.conf.SSLEnabled() {
		a.logger.Debug().Msg("SSL support enabled.")
		if a.conf.AutoTLS {
			return a.startAutoTLS(address)
		}

		return a.e.StartTLS(address,
			fmt.Sprintf("%s/%s", a.conf.CertCacheDir, a.conf.Hostname),
			fmt.Sprintf("%s/%s", a.conf.CertCacheDir, a.conf.Hostname))
	}

	return a.e.Start(address)
}

// Shutdown terminate the API server cleanly
func (a *API) Shutdown(ctx context.Context) error {
	a.logger.Debug().Msg("shutting down API.")
	return a.e.Shutdown(ctx)
}

func (a *API) startAutoTLS(address string) error {
	a.logger.Debug().Msg("starting API using auto TLS support.")
	// since we are using LetsEncrypt we can only use port 443
	parts := strings.Split(address, ":")
	if len(parts) == 2 {
		return a.e.StartAutoTLS(parts[0] + ":443")
	}

	return a.e.StartAutoTLS(address + ":443")
}
