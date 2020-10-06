package api

import (
	"github.com/creekorful/open-dydns/proto"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"time"
)

// getAuthMiddleware instantiate a authentication middleware
func getAuthMiddleware(signingKey string) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: []byte(signingKey),
	})
}

// getUserContext extract the user context from current request
func getUserContext(c echo.Context) proto.UserContext {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	return proto.UserContext{
		UserID: uint(claims["userID"].(float64)),
	}
}

// makeToken create & signed a new JWT token
func makeToken(userCtx proto.UserContext, secretKey string, tokenTTL time.Duration) (proto.TokenDto, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = userCtx.UserID

	if tokenTTL != 0 {
		claims["exp"] = time.Now().Add(tokenTTL).Unix()
	}

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return proto.TokenDto{}, err
	}

	return proto.TokenDto{
		Token: t,
	}, nil
}
