package service

import (
	"context"
	"errors"
	"github.com/tiagods/auth/internal/infra/logger"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/gommon/log"
	"github.com/tiagods/auth/internal/adapter/database"
	"github.com/tiagods/auth/internal/adapter/web/presenter/request"
	"github.com/tiagods/auth/internal/adapter/web/presenter/response"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/cache"
	"github.com/tiagods/auth/internal/infra/cripto"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/message"
	"github.com/tiagods/auth/internal/infra/tracer"
)

type (
	tokenService struct {
		repo  database.Repository
		cache cache.Repository
	}

	TokenService interface {
		Register(ctx context.Context, register *request.Register) (response.Register, error)
		Login(ctx context.Context, login *request.Login) (response.Token, error)
		RecreateToken(ctx context.Context, tokenReq *request.RefreshToken) (response.Token, error)
		RevokeToken(ctx context.Context, token *request.RefreshToken) error
		ValidateToken(ctx context.Context, authorization string) error
	}
)

func NewTokenService(repo database.Repository, cache cache.Repository) TokenService {
	return &tokenService{
		repo, cache,
	}
}

func (t *tokenService) ValidateToken(ctx context.Context, authorization string) error {
	caller := "service::validate_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	if authorization == "" {
		msg := message.ErrLoginRequired
		tracer.SetError(span, msg.GetError())
		return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	authorization = strings.ReplaceAll(authorization, "Bearer ", "")

	_, claims, err := t.parseToken(ctx, authorization)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}

	user, err := t.getUser(ctx, claims)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}
	if _, err = t.getToken(ctx, user.ID, authorization); err != nil {
		tracer.SetError(span, err)
		logger.Error(ctx, err, "failed to get token from cache")
		return err

	}
	return nil
}

func (t *tokenService) Register(ctx context.Context, register *request.Register) (response.Register, error) {
	caller := "service::register"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	pass, err := cripto.Encode(register.Password)
	if err != nil {
		tracer.SetError(span, err)
		return response.Register{}, err
	}

	user := &entity.UserCredential{Username: register.Username, Password: pass}
	err = t.repo.RegisterAccount(ctx, nil, user)
	if err != nil {
		tracer.SetError(span, err)
		return response.Register{}, err
	}
	return response.Register{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

func (t *tokenService) Login(ctx context.Context, login *request.Login) (response.Token, error) {
	caller := "service::login"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	pass, err := cripto.Encode(login.Password)
	if err != nil {
		tracer.SetError(span, err)
		return response.Token{}, err
	}

	result, err := t.repo.FindByUserAndPassword(ctx, login.Username, pass)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			msg := message.ErrUserNotFound
			logger.Warn(ctx, err, msg.UserMessage)
			tracer.SetWarning(span, msg.GetError())
			return response.Token{}, httperrors.NewHttpError(ctx, http.StatusBadRequest, msg, msg.GetError())
		}
		tracer.SetError(span, err)
		return response.Token{}, err
	}
	user := &entity.User{ID: result.ID, Username: result.Username}

	token, err := t.generateTokenPair(ctx, user, false)
	if err != nil {
		tracer.SetError(span, err)
		logger.Error(ctx, err, "failed to generate token pair")
		return response.Token{}, err
	}

	return token, nil
}

func (t *tokenService) RevokeToken(ctx context.Context, token *request.RefreshToken) error {
	caller := "service::revoke_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	if token.RefreshToken != "" {
		rt := &entity.RefreshToken{}
		err := t.cache.Get(ctx, entity.RefreshToken{ID: token.RefreshToken}.GetKey(), rt)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				tracer.SetWarning(span, err)
				msg := message.ErrRefreshNotFound
				return httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
			}
			tracer.SetError(span, err)
			return err
		}
		err = t.repo.DeleteRefreshToken(ctx, nil, rt.UserID)
		if err != nil {
			tracer.SetError(span, err)
			logger.Error(ctx, err, "failed to delete refresh token")
			return err
		}
		err = t.cache.Delete(ctx, rt.GetKey())
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				tracer.SetWarning(span, err)
				logger.Warn(ctx, err, "refresh token not found in cache")
				return nil // ID already deleted, no error needed
			}
			tracer.SetError(span, err)
			logger.Error(ctx, err, "failed to delete refresh token from cache")
			return err
		}

	}
	return httperrors.NewHttpError(ctx, http.StatusBadRequest, message.ErrRefreshNotFound, nil)
}

func (t *tokenService) RecreateToken(ctx context.Context, tokenReq *request.RefreshToken) (response.Token, error) {
	caller := "service::refresh_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	token, claims, err := t.parseToken(ctx, tokenReq.RefreshToken)
	if err != nil {
		tracer.SetError(span, err)
		return response.Token{}, err
	}
	usr, err := t.getUser(ctx, claims)
	if err != nil {
		tracer.SetError(span, err)
		return response.Token{}, err
	}
	refreshToken, err := t.getRefreshToken(ctx, usr.ID, token.Raw)
	if err != nil {
		return response.Token{}, err
	}
	if refreshToken == nil {
		logger.Warn(ctx, nil, "Refresh token not found in cache or repository")
		msg := message.ErrRefreshNotFound
		tracer.SetWarning(span, msg.GetError())
		return response.Token{}, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}
	return t.generateTokenPair(ctx, usr, true)

}

func (t *tokenService) generateTokenPair(ctx context.Context, user *entity.User, updateToken bool) (response.Token, error) {
	caller := "service::generate_token_pair"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	tk := entity.Token{UserID: user.ID}
	isGenerateToken := updateToken

	result := &entity.Token{}
	err := t.cache.Get(ctx, tk.GetKey(), result)
	if errors.Is(err, cache.ErrNotFound) {
		isGenerateToken = true
	} else {
		tk.ID = result.ID
	}

	if isGenerateToken {
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["sub"] = 1
		claims["name"] = user.Username
		claims["user_id"] = user.ID
		claims["admin"] = true
		exp := time.Now().Add(time.Second * 120)
		claims["exp"] = exp.Unix()

		signature, err := token.SignedString([]byte("secret"))
		if err != nil {
			tracer.SetError(span, err)
			return response.Token{}, err
		}

		tk.ID = signature

		err = t.cache.Set(ctx, tk.GetKey(), tk, time.Until(exp))
		if err != nil {
			tracer.SetError(span, err)
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
			if errors.Is(err, database.ErrNotFound) {
				tracer.SetWarning(span, err)
				rk = nil
			} else {
				tracer.SetError(span, err)
				return response.Token{}, err
			}
		}
		// Agora verifica rk explicitamente se necessário
		if rk == nil {
			isGenerateRefreshToken = true
		} else {
			refresh.ID = rk.ID
			err = t.cache.Set(ctx, refresh.GetKey(), refresh, time.Until(rk.ExpiresAt.Local()))
			if err != nil {
				tracer.SetError(span, err)
				return response.Token{}, err
			}
		}
	} else if rsRefresh.ID == "" {
		isGenerateRefreshToken = true
	} else {
		refresh.ID = rsRefresh.ID
	}

	if isGenerateRefreshToken {
		exp := time.Now().Add(time.Hour * 2)
		//claims := jwt.RegisteredClaims{
		//	ExpiresAt: jwt.NewNumericDate(exp),
		//}

		refreshToken := jwt.New(jwt.SigningMethodHS256)

		rtClaims := refreshToken.Claims.(jwt.MapClaims)
		rtClaims["sub"] = 1
		rtClaims["name"] = user.Username
		rtClaims["user_id"] = user.ID
		rtClaims["exp"] = exp.Unix()

		signature, err := refreshToken.SignedString([]byte("secret"))
		if err != nil {
			tracer.SetError(span, err)
			return response.Token{}, err
		}
		refresh.ID = signature

		err = t.cache.Set(ctx, refresh.GetKey(), refresh, time.Until(exp))
		if err != nil {
			tracer.SetError(span, err)
			return response.Token{}, err
		}

		tx, err := t.repo.BeginTransaction()
		if err != nil {
			return response.Token{}, err
		}
		err = t.repo.UpdateRefreshToken(ctx, tx, refresh.UserID, refresh.ID, exp)
		if err != nil {
			tracer.SetError(span, err)
			errRollback := tx.Rollback()
			if errRollback != nil {
				log.Error(errRollback)
			}
			return response.Token{}, err
		}
		err = tx.Commit()
		if err != nil {
			tracer.SetError(span, err)
			return response.Token{}, err
		}
	}

	return response.Token{
		AccessToken:  tk.ID,
		RefreshToken: refresh.ID,
	}, nil
}
