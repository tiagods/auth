package database

import (
	"context"
	"errors"
	"net/http"

	"github.com/tiagods/auth/internal/adapter/database/model"
	"github.com/tiagods/auth/internal/infra/httperrors"
)

type (
	memoryRepository struct{}
)

var users []model.User
var tokens []model.RefreshToken

func NewMemoryRepository() Repository {
	var users []model.User
	tokens = make([]model.RefreshToken, 0)

	users = append(users, model.User{
		ID:       "jon",
		Username: "jon",
		Password: "password",
	})

	users = append(users, model.User{
		ID:       "tiago",
		Username: "tiago",
		Password: "password",
	})

	return memoryRepository{}
}

func (m memoryRepository) RegisterAccount(ctx context.Context, user model.User) error {
	for _, value := range users {
		if value.Username == user.Username {
			err := errors.New("user already registered")
			return httperrors.NewHttpError(http.StatusConflict, err.Error(), err)
		}
	}
	users = append(users, user)
	return nil
}

func (m memoryRepository) FindByUserAndPassword(ctx context.Context, username string, password string) (model.User, error) {
	for _, usr := range users {
		if usr.Username == username &&
			usr.Password == password {
			return usr, nil
		}
	}
	err := errors.New("user not found")
	return model.User{}, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
}

func (m memoryRepository) UpdateRefreshToken(ctx context.Context, userId string, newToken string) error {
	found := false
	for i, tk := range tokens {
		if tk.ID == userId {
			tokens[i].RefreshToken = newToken
			found = true
			break
		}
	}
	if !found {
		tokens = append(tokens, model.RefreshToken{ID: userId, RefreshToken: newToken})
	}
	return nil
}

func (m memoryRepository) FindRefreshToken(ctx context.Context, refreshToken string) (model.User, error) {
	for _, tk := range tokens {
		if tk.RefreshToken == refreshToken {
			for _, us := range users {
				if us.ID == tk.ID {
					return us, nil
				}
			}
		}
	}
	err := errors.New("not authorized refresh token")
	return model.User{}, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
}
