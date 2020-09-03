package api

import (
	"context"
	"github.com/creekorful/open-dydns/internal/opendydnsd/daemon"
	"github.com/creekorful/open-dydns/internal/proto"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
)

type Api struct {
	e          *echo.Echo
	signingKey []byte
}

func NewAPI(d daemon.Daemon, signingKey string) (*Api, error) {
	// Configure echo
	e := echo.New()
	e.HideBanner = false
	e.Logger.SetOutput(ioutil.Discard)

	// Create the API
	a := Api{
		e:          e,
		signingKey: []byte(signingKey),
	}

	// Register global middlewares
	e.Use(newZeroLogMiddleware())

	// Register per-route middlewares
	authMiddleware := getAuthMiddleware(a.signingKey)

	// Register endpoints
	e.POST("/sessions", a.Authenticate(d))
	e.GET("/aliases", a.GetAliases(d), authMiddleware)
	e.POST("/aliases", a.RegisterAlias(d), authMiddleware)
	e.DELETE("/aliases/{name}", a.DeleteAlias(d), authMiddleware)

	return &a, nil
}

func (a *Api) Authenticate(d daemon.Daemon) echo.HandlerFunc {
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

func (a *Api) GetAliases(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		aliases, err := d.GetAliases(userCtx)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, aliases)
	}
}

func (a *Api) RegisterAlias(d daemon.Daemon) echo.HandlerFunc {
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

func (a *Api) DeleteAlias(d daemon.Daemon) echo.HandlerFunc {
	return func(c echo.Context) error {
		userCtx := getUserContext(c)

		alias := c.Param("name")

		if err := d.DeleteAlias(userCtx, alias); err != nil {
			return err
		}

		return c.NoContent(http.StatusOK)
	}
}

func (a *Api) Start(address string) error {
	return a.e.Start(address)
}

func (a *Api) Shutdown(ctx context.Context) error {
	return a.e.Shutdown(ctx)
}
