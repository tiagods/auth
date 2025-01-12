package repository

import (
	"context"
	"database/sql"
	"github.com/tiagods/auth/internal/domain/entity"
	"time"
)

type (
	UserRepository interface {
		Ping() error
		BeginTransaction() (*sql.Tx, error)
		RegisterAccount(ctx context.Context, tx *sql.Tx, user entity.User) error
		GetRefreshToken(ctx context.Context, userID int64, refreshToken *string) (*entity.RefreshToken, error)
		UpdateRefreshToken(ctx context.Context, tx *sql.Tx, userId int64, newToken string, expiresAt time.Time) error
		FindByUserAndPassword(ctx context.Context, username string, password string) (*entity.User, error)
	}
)
