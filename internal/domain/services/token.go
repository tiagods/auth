package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/gommon/log"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"net/http"
	"time"
)

type (
	tokenService struct {
		repo  database.Repository
		cache cache.Repository
	}

	TokenService interface {
	}
)

func NewTokenService(repo database.Repository, cache cache.Repository) TokenService {
	return &tokenService{
		repo:  repo,
		cache: cache,
	}
}
func (t tokenService) Login(ctx context.Context, login *request.Login) (response.Token, error) {
	result, err := t.repo.FindByUserAndPassword(ctx, login.Username, login.Password)
	if err != nil {
		return response.Token{}, err
	}

	return response.Token{}, nil
}

func RefreshToken(tokenReq *request.RefreshToken) (response.Token, error) {
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

	errLoginRequired := errors.New("login required")

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if _, ok := claims["sub"].(float64); ok {
			if int(claims["sub"].(float64)) == 1 {
				id, rtOk := refreshTokens[token.Raw]
				if rtOk {
					for i, usr := range users {
						if usr.ID == id {
							usr.TokenEx = nil
							newTokenPair, err := generateTokenPair(i, usr)
							if err != nil {
								log.Error(err)
								return response.Token{}, err
							}
							return newTokenPair, nil
						}
					}
					log.Warn(err)
					return response.Token{}, httperrors.NewHttpError(http.StatusUnauthorized, errLoginRequired.Error(), errLoginRequired)
				}
				return response.Token{}, httperrors.NewHttpError(http.StatusUnauthorized, errLoginRequired.Error(), errLoginRequired)
			}
		}
		return response.Token{}, httperrors.NewHttpError(http.StatusUnauthorized, errLoginRequired.Error(), errLoginRequired)
	}
	return response.Token{}, err
}

type User struct {
	ID             string
	Username       string
	Password       string
	Token          string
	TokenEx        *time.Time
	RefreshToken   string
	RefreshTokenEx *time.Time
}

var users []User
var refreshTokens = map[string]string{}

func generateTokenPair(user User) (response.Token, error) {
	now := time.Now()
	if user.Token == "" || user.TokenEx == nil || now.After(*user.TokenEx) {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["sub"] = 1
		claims["name"] = user.Username
		claims["admin"] = true
		exp := time.Now().Add(time.Second * 30)
		claims["exp"] = exp

		t, err := token.SignedString([]byte("secret"))
		if err != nil {
			return response.Token{}, err
		}
		user.Token = t
		user.TokenEx = &exp
	}

	if user.RefreshToken == "" || user.RefreshTokenEx == nil || now.After(*user.RefreshTokenEx) {
		refreshToken := jwt.New(jwt.SigningMethodHS256)
		rtClaims := refreshToken.Claims.(jwt.MapClaims)
		rtClaims["sub"] = 1
		exp := time.Now().Add(time.Hour * 24)
		rtClaims["exp"] = exp

		rt, err := refreshToken.SignedString([]byte("secret"))
		if err != nil {
			return response.Token{}, err
		}
		user.RefreshToken = rt
		user.RefreshTokenEx = &exp

		refreshTokens[rt] = user.ID
	}

	return response.Token{
		AccessToken:  user.Token,
		RefreshToken: user.RefreshToken,
	}, nil
}
