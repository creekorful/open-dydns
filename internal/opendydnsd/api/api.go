package api

import (
	"context"
	"github.com/creekorful/open-dydns/internal/opendydnsd/config"
	"github.com/creekorful/open-dydns/internal/opendydnsd/daemon"
	"github.com/creekorful/open-dydns/pkg/proto"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

// API represent the Daemon REST API
type API struct {
	e          *echo.Echo
	signingKey []byte
}

// NewAPI return a new API instance, wrapped around given Daemon instance
// and with given config
func NewAPI(d daemon.Daemon, conf config.APIConfig) (*API, error) {
	// Configure echo
	e := echo.New()
	e.HideBanner = false
	e.Logger.SetOutput(ioutil.Discard)

	// Create the API
	a := API{
		e:          e,
		signingKey: []byte(conf.SigningKey),
	}

	// Register global middlewares
	e.Use(newZeroLogMiddleware(d.Logger()))

	// Register per-route middlewares
	authMiddleware := getAuthMiddleware(a.signingKey)

	// Register endpoints
	e.POST("/sessions", a.authenticate(d))
	e.GET("/aliases", a.getAliases(d), authMiddleware)
	e.POST("/aliases", a.registerAlias(d), authMiddleware)
	e.PUT("/aliases", a.updateAlias(d), authMiddleware)
	e.DELETE("/aliases/:name", a.deleteAlias(d), authMiddleware)

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
		token, err := makeToken(userCtx, a.signingKey)
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

// Start start the API server
func (a *API) Start(address string) error {
	return a.e.Start(address)
}

// Shutdown terminate the API server cleanly
func (a *API) Shutdown(ctx context.Context) error {
	return a.e.Shutdown(ctx)
}
