package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/tiagods/auth/internal/adapter/database/model"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"github.com/tiagods/auth/internal/infra/message"
	"net/http"
	"time"
)

type (
	memoryRepository struct{}
)

var users []model.User
var tokens map[int64]model.RefreshToken

func NewMemoryRepository() Repository {
	var users []model.User
	tokens = make(map[int64]model.RefreshToken)

	users = append(users, model.User{
		ID:       1,
		Username: "jon",
		Password: "password",
	})

	users = append(users, model.User{
		ID:       2,
		Username: "tiago",
		Password: "password",
	})

	return memoryRepository{}
}

func (m memoryRepository) Ping() error {
	return nil
}

func (m memoryRepository) BeginTransaction() (*sql.Tx, error) {
	return nil, nil
}

func (m memoryRepository) GetRefreshToken(ctx context.Context, userID int64, refreshToken *string) (*entity.RefreshToken, error) {
	for _, tk := range tokens {
		if userID != 0 && refreshToken != nil {
			if tk.UserID == userID && tk.ID == *refreshToken {
				return tk.ToEntity(), nil
			}
		} else {
			if userID != 0 && tk.UserID == userID {
				return tk.ToEntity(), nil
			}
			if refreshToken != nil && tk.ID == *refreshToken {
				return tk.ToEntity(), nil
			}
		}
	}

	return nil, nil
}

func (m memoryRepository) RegisterAccount(ctx context.Context, tx *sql.Tx, user entity.User) error {
	for _, value := range users {
		if value.Username == user.Username {
			msg := message.ErrDuplicateUser
			err := errors.New(msg.UserMessage)
			return httperrors.NewHttpError(ctx, http.StatusConflict, msg, err)
		}
	}

	id := len(users) + 1
	resultUser := model.User{
		ID:       int64(id),
		Username: user.Username,
		Password: "password",
	}

	users = append(users, resultUser)
	return nil
}

func (m memoryRepository) FindByUserAndPassword(ctx context.Context, username string, password string) (*entity.User, error) {
	for _, usr := range users {
		if usr.Username == username &&
			usr.Password == password {
			return usr.ToEntity(), nil
		}
	}
	msg := message.ErrUserNotFound
	return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
}

func (m memoryRepository) UpdateRefreshToken(ctx context.Context, tx *sql.Tx, userId int64, newToken string, expiresAt time.Time) error {
	if tk, ok := tokens[userId]; ok {
		tk.ID = newToken
		tk.ExpiresAt = expiresAt
		tokens[userId] = tk
	} else {
		tokens[userId] = model.RefreshToken{UserID: userId, ID: newToken, ExpiresAt: expiresAt, CreatedAt: time.Now().UTC()}
	}
	return nil
}
