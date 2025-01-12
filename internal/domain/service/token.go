package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/message"

	"net/http"
	"time"
)

type (
	tokenService struct {
		repo  database.Repository
		cache cache.Repository
	}

	TokenService interface {
		Login(ctx context.Context, login *request.Login) (response.Token, error)
		RefreshToken(ctx context.Context, tokenReq *request.RefreshToken) (response.Token, error)
	}
)

func NewTokenService(repo database.Repository, cache cache.Repository) TokenService {
	return &tokenService{
		repo, cache,
	}
}

func (t *tokenService) Login(ctx context.Context, login *request.Login) (response.Token, error) {
	result, err := t.repo.FindByUserAndPassword(ctx, login.Username, login.Password)
	if err != nil {
		return response.Token{}, err
	}
	user := &entity.User{ID: result.ID, Username: result.Username}

	token, err := t.generateTokenPair(ctx, user, false)
	if err != nil {
		return response.Token{}, err
	}
	return token, nil
}

func (t *tokenService) RefreshToken(ctx context.Context, tokenReq *request.RefreshToken) (response.Token, error) {
	token, err := jwt.Parse(tokenReq.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			msg := message.ErrLoginRequired
			return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, err)
		}
		return []byte("secret"), nil
	})

	if token == nil {
		msg := message.ErrInvalidToken
		return response.Token{}, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if _, ok := claims["sub"].(float64); ok {
			if int(claims["sub"].(float64)) == 1 {
				refreshTokenString := token.Raw
				token, err := t.updateToken(ctx, entity.RefreshToken{ID: refreshTokenString})
				if err != nil {
					return response.Token{}, err
				}
				return token, nil
			}
		}
		msg := message.ErrLoginRequired
		return response.Token{}, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}
	return response.Token{}, err
}

func (t *tokenService) updateToken(ctx context.Context, refreshToken entity.RefreshToken) (response.Token, error) {
	user, err := t.getTokenByRefresh(ctx, refreshToken)
	if err != nil {
		return response.Token{}, err
	}

	return t.generateTokenPair(ctx, user, true)
}

func (t *tokenService) getTokenByRefresh(ctx context.Context, refreshToken entity.RefreshToken) (*entity.User, error) {
	rs, err := t.repo.GetRefreshToken(ctx, refreshToken.UserID, &refreshToken.ID)
	if err != nil {
		return nil, err
	}
	usr := &entity.User{ID: rs.UserID}

	refreshToken.UserID = usr.ID

	notfound := cache.ErrNotFound
	err = t.cache.Get(ctx, refreshToken.ID, &entity.User{})
	if err != nil {
		if errors.Is(err, notfound) {
			msg := message.ErrLoginRequired
			return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
		}
		return nil, err
	}
	return usr, nil
}

func (t *tokenService) generateTokenPair(ctx context.Context, user *entity.User, updateToken bool) (response.Token, error) {
	tk := entity.Token{UserID: user.ID}
	isGenerateToken := updateToken

	result := &entity.Token{}
	err := t.cache.Get(ctx, tk.GetKey(), result)
	if errors.Is(err, cache.ErrNotFound) {
		isGenerateToken = true
	} else {
		tk.Token = result.Token
	}

	if isGenerateToken {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["sub"] = 1
		claims["name"] = user.Username
		claims["admin"] = true
		exp := time.Now().Add(time.Second * 30)
		claims["exp"] = exp

		signature, err := token.SignedString([]byte("secret"))
		if err != nil {
			return response.Token{}, err
		}

		tk.Token = signature

		err = t.cache.Set(ctx, tk.GetKey(), tk, time.Second*30)
		if err != nil {
			return response.Token{}, err
		}
	}

	refresh := &entity.RefreshToken{UserID: user.ID}
	isGenerateRefreshToken := false
	rsRefresh := &entity.RefreshToken{}
	err = t.cache.Get(ctx, refresh.GetKey(), rsRefresh)
	if errors.Is(err, cache.ErrNotFound) {
		rk, err := t.repo.GetRefreshToken(ctx, user.ID, nil)
		if err != nil {
			return response.Token{}, err
		}
		if rk.ID == "" {
			isGenerateRefreshToken = true
		} else {
			refresh.ID = rk.ID
			err = t.cache.Set(ctx, refresh.GetKey(), refresh, time.Hour*2)
			if err != nil {
				return response.Token{}, err
			}
		}
	} else {
		refresh.ID = rsRefresh.ID
	}

	if isGenerateRefreshToken {
		refreshToken := jwt.New(jwt.SigningMethodHS256)
		rtClaims := refreshToken.Claims.(jwt.MapClaims)
		rtClaims["sub"] = 1
		exp := time.Now().Add(time.Hour * 2)
		rtClaims["exp"] = exp

		signature, err := refreshToken.SignedString([]byte("secret"))
		if err != nil {
			return response.Token{}, err
		}
		refresh.ID = signature

		err = t.cache.Set(ctx, refresh.GetKey(), tk, time.Hour*2)
		if err != nil {
			return response.Token{}, err
		}

		tx, err := t.repo.BeginTransaction()
		if err != nil {
			return response.Token{}, err
		}
		err = t.repo.UpdateRefreshToken(ctx, tx, refresh.UserID, refresh.ID, exp)
		if err != nil {
			errRollback := tx.Rollback()
			if errRollback != nil {
				log.Error(errRollback)
			}
			return response.Token{}, err
		}
		err = tx.Commit()
		if err != nil {
			return response.Token{}, err
		}
	}

	return response.Token{
		AccessToken:  tk.Token,
		RefreshToken: refresh.ID,
	}, nil
}
