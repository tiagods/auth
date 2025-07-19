package middleware

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt"
	"github.com/labstack/echo/v4"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/message"
	"github.com/tiagods/auth/internal/infra/requestcontext"
	"github.com/tiagods/auth/internal/infra/utils"
	"net/http"
	"strings"
	"time"
)

var secret = []byte("secret")

var IsLoggedIn = echojwt.WithConfig(echojwt.Config{
	SigningKey: secret,
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

func NewValidationMiddleware(cacheRepo cache.Repository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return validateToken(cacheRepo, next)
	}
}

func validateToken(cacheRepo cache.Repository, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		authorization := c.Request().Header.Get("Authorization")
		if strings.HasPrefix(authorization, "Bearer ") {
			authorization = strings.TrimPrefix(authorization, "Bearer ")
		}

		token, err := jwt.Parse(authorization, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

		if err != nil {
			if exp, ok := token.Claims.(jwt.MapClaims)["exp"]; ok && exp != nil {
				if exp.(float64) > float64(time.Now().Unix()) {
					msg := message.ErrLoginRequired
					logger.Error(ctx, err, fmt.Sprintf("ID expired at: %v", time.Unix(int64(exp.(float64)), 0)))
					return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
				}
				logger.Info(ctx, fmt.Sprintf("ID will expired at: %v", time.Unix(int64(exp.(float64)), 0)))
			} else {
				logger.Error(ctx, err, fmt.Sprintf("Error parsing token: %v", token))
				msg := message.ErrInvalidToken
				return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
			}
		}

		if !token.Valid {
			logger.Error(ctx, nil, fmt.Sprintf("Invalid token: %v", token.Claims))
			msg := message.ErrInvalidToken
			return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error(ctx, nil, "Invalid token claims")
			msg := message.ErrInvalidToken
			return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
		}

		id, ok := claims["user_id"].(float64)
		if !ok || id <= 0 {
			msg := message.ErrInvalidToken
			return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
		}
		name, ok := claims["name"].(string)
		if !ok || name == "" {
			msg := message.ErrInvalidToken
			return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
		}

		foundTk := &entity.Token{}
		err = cacheRepo.Get(ctx, entity.Token{ID: token.Raw, UserID: int64(id)}.GetKey(), foundTk)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				logger.Error(ctx, err, "ID not found in cache")
				msg := message.ErrLoginRequired
				return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
			} else {
				logger.Error(ctx, err, "Error retrieving token from cache")
				return err
			}
		}
		if foundTk.ID != token.Raw {
			logger.Error(ctx, nil, "Token ID does not match")
			msg := message.ErrInvalidToken
			return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
		}

		rc := utils.GetRequestContext(c)
		rc.UserID = int64(id)
		rc.UserName = name

		ctx = context.WithValue(ctx, requestcontext.ContextKey, rc)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)

	}
}
