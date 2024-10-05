package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/httperrors"
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

var errLoginRequired = errors.New("login required")

func NewTokenService() *tokenService {
	return &tokenService{}
}

func (t *tokenService) WithRepository(repo database.Repository) *tokenService {
	t.repo = repo
	return t
}

func (t *tokenService) WithCache(cache cache.Repository) *tokenService {
	t.cache = cache
	return t
}

func (t *tokenService) Login(ctx context.Context, login *request.Login) (response.Token, error) {
	result, err := t.repo.FindByUserAndPassword(ctx, login.Username, login.Password)
	if err != nil {
		return response.Token{}, err
	}
	user := &entity.User{ID: result.ID, Username: result.Username}
	token, err := t.generateTokenPair(user, false)
	if err != nil {
		return response.Token{}, err
	}
	return token, nil
}

func (t *tokenService) RefreshToken(ctx context.Context, tokenReq *request.RefreshToken) (response.Token, error) {
	token, err := jwt.Parse(tokenReq.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			log.Error(err)
			return nil, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
		}
		return []byte("secret"), nil
	})

	if token == nil {
		err = errors.New("invalid refresh RefreshToken")
		log.Error(err)
		return response.Token{}, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if _, ok := claims["sub"].(float64); ok {
			if int(claims["sub"].(float64)) == 1 {
				refreshTokenString := token.Raw
				token, err := t.updateToken(ctx, entity.RefreshToken{RefreshToken: refreshTokenString})
				if err != nil {
					return response.Token{}, err
				}
				return token, nil
			}
		}
		return response.Token{}, httperrors.NewHttpError(http.StatusUnauthorized, errLoginRequired.Error(), errLoginRequired)
	}
	return response.Token{}, err
}

func (t *tokenService) updateToken(ctx context.Context, refreshToken entity.RefreshToken) (response.Token, error) {
	user, err := t.getTokenByRefresh(ctx, refreshToken)
	if err != nil {
		return response.Token{}, err
	}

	return t.generateTokenPair(user, true)
}

func (t *tokenService) getTokenByRefresh(ctx context.Context, refreshToken entity.RefreshToken) (*entity.User, error) {
	rs, err := t.repo.FindRefreshToken(ctx, refreshToken.RefreshToken)
	if err != nil {
		return nil, err
	}
	usr := &entity.User{ID: rs.ID, Username: rs.Username}

	refreshToken.UserID = usr.ID

	notfound := cache.ErrNotFound
	_, err = t.cache.Get(refreshToken.RefreshToken)
	if err != nil {
		if errors.Is(err, notfound) {
			return nil, httperrors.NewHttpError(http.StatusUnauthorized, errLoginRequired.Error(), errLoginRequired)
		}
		return nil, err
	}
	return usr, nil
}

func (t *tokenService) generateTokenPair(user *entity.User, updateToken bool) (response.Token, error) {
	sr := entity.Token{UserID: user.ID}
	notfound := cache.ErrNotFound

	isGenerateToken := updateToken
	rs, err := t.cache.Get(sr.GetKey())
	if errors.Is(err, notfound) {
		isGenerateToken = true
	} else {
		token, ok := rs.(*entity.Token)
		if !ok {
			isGenerateToken = true
		} else {
			sr.Token = token.Token
		}
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

		sr.Token = signature

		err = t.cache.Set(sr.GetKey(), sr, time.Second*30)
		if err != nil {
			return response.Token{}, err
		}
	}

	refresh := &entity.RefreshToken{UserID: user.ID}
	isGenerateRefreshToken := false
	rsRefresh, err := t.cache.Get(refresh.GetKey())
	if errors.Is(err, notfound) {
		isGenerateRefreshToken = true
	} else {
		tokenRef, ok := rsRefresh.(*entity.RefreshToken)
		if !ok {
			isGenerateRefreshToken = true
		} else {
			refresh.RefreshToken = tokenRef.RefreshToken
		}
	}

	if isGenerateRefreshToken {
		refreshToken := jwt.New(jwt.SigningMethodHS256)
		rtClaims := refreshToken.Claims.(jwt.MapClaims)
		rtClaims["sub"] = 1
		exp := time.Now().Add(time.Hour * 24)
		rtClaims["exp"] = exp

		signature, err := refreshToken.SignedString([]byte("secret"))
		if err != nil {
			return response.Token{}, err
		}
		refresh.RefreshToken = signature

		err = t.cache.Set(refresh.GetKey(), sr, time.Hour*24)
		if err != nil {
			return response.Token{}, err
		}

		err = t.repo.UpdateRefreshToken(context.Background(), refresh.UserID, refresh.RefreshToken)
		if err != nil {
			return response.Token{}, err
		}
	}

	return response.Token{
		AccessToken:  sr.Token,
		RefreshToken: refresh.RefreshToken,
	}, nil
}
