package token_middleware

import (
	"github.com/dgrijalva/jwt-go"
	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4"
	"net/http"
)

var IsLoggedIn = echojwt.WithConfig(echojwt.Config{
	SigningKey: []byte("secret"),
})

func IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		isAdmin := claims["admin"].(bool)

		if !isAdmin {
			return echo.ErrUnauthorized
		}

		return next(c)
	}
}

func Private(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.String(http.StatusOK, "Welcome "+name+"!")
}
