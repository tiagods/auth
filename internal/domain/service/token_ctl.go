package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/logger"
	"github.com/tiagods/auth/internal/infra/message"
	"github.com/tiagods/auth/internal/infra/tracer"
	"net/http"
	"time"
)

func (t *tokenService) parseToken(ctx context.Context, tokenReq string) (*jwt.Token, jwt.MapClaims, error) {
	caller := "service::parse_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	//token, err := jwt.Parse(tokenReq.RecreateToken, func(token *jwt.ID) (interface{}, error) {
	//	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	//		exp := token.Claims.(jwt.MapClaims)["exp"]
	//		err := fmt.Errorf("unexpected signing method: %v ,expiration: %v", token.Header["alg"], exp)
	//		msg := message.ErrLoginRequired
	//		tracer.SetError(span, err)
	//		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, err)
	//	}
	//	return []byte("secret"), nil
	//})

	token, err := jwt.Parse(tokenReq, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if err != nil {
		if exp, ok := token.Claims.(jwt.MapClaims)["exp"]; ok && exp != nil {
			logger.Error(ctx, err, fmt.Sprintf("ID expired at: %v", time.Unix(int64(exp.(float64)), 0)))
		}

		logger.Error(ctx, err, fmt.Sprintf("Error parsing token: %v", tokenReq))
		msg := message.ErrInvalidToken
		return nil, nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	if token == nil {
		msg := message.ErrInvalidToken
		tracer.SetError(span, err)
		return nil, nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		msg := message.ErrInvalidToken
		tracer.SetError(span, err)
		return nil, nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	_, err = t.getUser(ctx, claims)
	if err != nil {
		return nil, nil, err
	}

	if !token.Valid {
		logger.Error(ctx, nil, fmt.Sprintf("Invalid token: %v", token.Claims))
		msg := message.ErrRefreshNotFound
		tracer.SetError(span, msg.GetError())
		return nil, nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	return token, claims, nil
}

func (t *tokenService) getUser(ctx context.Context, claims jwt.MapClaims) (*entity.User, error) {
	caller := "service::get_user"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	id, ok := claims["user_id"].(float64)
	if !ok || id <= 0 {
		msg := message.ErrInvalidToken
		tracer.SetError(span, fmt.Errorf("invalid refresh token, user_id not found"))
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}
	name, ok := claims["name"].(string)
	if !ok || name == "" {
		msg := message.ErrInvalidToken
		tracer.SetError(span, fmt.Errorf("invalid refresh token, name not found"))
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}
	return &entity.User{ID: int64(id), Username: name}, nil
}

func (t *tokenService) getToken(ctx context.Context, userID int64, tokenString string) (*entity.Token, error) {
	tk := &entity.Token{ID: tokenString, UserID: userID}
	err := t.cache.Get(ctx, tk.GetKey(), tk)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			logger.Warn(ctx, err, "token not found in cache")
			return nil, nil
		} else {
			logger.Error(ctx, err, "failed to get token from cache")
			return nil, err
		}
	}
	if tk.ID != tokenString {
		msg := message.ErrInvalidToken
		logger.Warn(ctx, nil, msg.UserMessage)
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}
	return tk, err
}

func (t *tokenService) getRefreshToken(ctx context.Context, userID int64, refreshTokenString string) (*entity.RefreshToken, error) {
	refreshToken := &entity.RefreshToken{ID: refreshTokenString, UserID: userID}
	foundRepo := false
	err := t.cache.Get(ctx, refreshToken.GetKey(), refreshToken)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			logger.Warn(ctx, err, "refresh token not found in cache")
			return nil, nil
		} else {
			logger.Error(ctx, err, "failed to get refresh token from cache")
		}
		refreshToken, err = t.repo.GetRefreshToken(ctx, userID, &refreshTokenString)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				logger.Warn(ctx, err, "refresh token not found in repository")
				return nil, nil
			} else {
				logger.Error(ctx, err, "failed to get refresh token from repository")
				return nil, err
			}
		}
		foundRepo = true
	}
	if refreshToken.ID != refreshTokenString {
		msg := message.ErrInvalidToken
		logger.Warn(ctx, nil, msg.UserMessage)
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}
	if foundRepo {
		err = t.cache.Set(ctx, refreshToken.GetKey(), refreshToken, time.Until(refreshToken.ExpiresAt))
		if err != nil {
			logger.Error(ctx, err, "failed to set refresh token in cache")
			return nil, err
		}
	}

	return refreshToken, nil
}
