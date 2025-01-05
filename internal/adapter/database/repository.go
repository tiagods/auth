package database

import (
	"context"
	"database/sql"
	"github.com/tiagods/auth/internal/domain/entity"

	"github.com/tiagods/auth/internal/adapter/database/model"
)

type (
	Repository interface {
		BeginTransaction() (*sql.Tx, error)
		RegisterAccount(ctx context.Context, user entity.User) error
		FindRefreshToken(ctx context.Context, refreshToken string) (model.User, error)
		UpdateRefreshToken(ctx context.Context, userId string, newToken string) error
		FindByUserAndPassword(ctx context.Context, username string, password string) (model.User, error)
	}

	repository struct {
		db *sql.DB
	}
)

func NewRepository(db *sql.DB) Repository {
	return &repository{db}
}

func (r *repository) BeginTransaction() (*sql.Tx, error) {
	return r.db.Begin()
}
