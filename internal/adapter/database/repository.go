package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/database"
	"time"
)

var ErrNotFound = errors.New("not found")

type (
	repository struct {
		reader *database.DbAdapter
		writer *database.DbAdapter
	}

	Repository interface {
		Ping() error
		BeginTransaction() (*sql.Tx, error)
		RegisterAccount(ctx context.Context, tx *sql.Tx, user *entity.UserCredential) error
		GetRefreshToken(ctx context.Context, userID int64, refreshToken *string) (*entity.RefreshToken, error)
		UpdateRefreshToken(ctx context.Context, tx *sql.Tx, userId int64, newToken string, expiresAt time.Time) error
		FindByUserAndPassword(ctx context.Context, username string, password string) (*entity.UserCredential, error)
		DeleteRefreshToken(ctx context.Context, tx *sql.Tx, userId int64) error
	}
)

func NewRepository(reader *database.DbAdapter, writer *database.DbAdapter) Repository {
	return &repository{reader: reader, writer: writer}
}

func (r *repository) BeginTransaction() (*sql.Tx, error) {
	return r.writer.Begin()
}

func (r *repository) Ping() error {
	err := r.reader.Ping()
	if err != nil {
		return err
	}
	return r.writer.Ping()
}
