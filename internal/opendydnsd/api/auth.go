package api

import (
	"github.com/creekorful/open-dydns/internal/proto"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func getAuthMiddleware(signingKey []byte) echo.MiddlewareFunc {
	return middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey: signingKey,
	})
}

func getUserContext(c echo.Context) proto.UserContext {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	return proto.UserContext{
		UserID: uint(claims["userID"].(float64)),
	}
}

func makeToken(userCtx proto.UserContext, secretKey []byte) (proto.TokenDto, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["userID"] = userCtx.UserID

	// Generate encoded token and send it as response.
	t, err := token.SignedString(secretKey)
	if err != nil {
		return proto.TokenDto{}, err
	}

	return proto.TokenDto{
		Token: t,
	}, nil
}
