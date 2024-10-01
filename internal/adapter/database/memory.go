package database

import (
	"context"
	"errors"
	"github.com/tiagods/auth/internal/adapter/database/entity"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"net/http"
	"sync"
)

type (
	memoryRepository struct {
		mu    sync.Mutex
		users []entity.User
	}
)

func NewMemoryRepository() Repository {
	var users []entity.User
	users = append(users, entity.User{
		ID:       "jon",
		Username: "jon",
		Password: "password",
	})

	users = append(users, entity.User{
		ID:       "tiago",
		Username: "tiago",
		Password: "password",
	})

	return memoryRepository{
		users: users,
	}
}

func (m memoryRepository) RegisterAccount(ctx context.Context, user entity.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, value := range m.users {
		if value.Username == user.Username {
			err := errors.New("user already registered")
			return httperrors.NewHttpError(http.StatusConflict, err.Error(), err)
		}
	}
	m.users = append(m.users, user)
	return nil
}

func (m memoryRepository) FindByUserAndPassword(ctx context.Context, username string, password string) (entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, usr := range m.users {
		if usr.Username == username &&
			usr.Password == password {
			return usr, nil
		}
	}
	err := errors.New("duplicated user found")
	return entity.User{}, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
}
