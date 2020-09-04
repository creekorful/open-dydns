package api

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func newZeroLogMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := next(c); err != nil {
				c.Error(err)
			}

			logger.Debug().
				Str("RemoteAddr", c.RealIP()).
				Int("Synchronize", c.Response().Status).
				Int64("Length", c.Response().Size).
				Msgf("%s %s", c.Request().Method, c.Path())
			return nil
		}
	}
}
