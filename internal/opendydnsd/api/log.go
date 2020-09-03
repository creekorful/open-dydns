package api

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

func newZeroLogMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := next(c); err != nil {
				c.Error(err)
			}

			log.Debug().
				Str("RemoteAddr", c.RealIP()).
				Int("Status", c.Response().Status).
				Int64("Length", c.Response().Size).
				Msgf("%s %s", c.Request().Method, c.Path())
			return nil
		}
	}
}
