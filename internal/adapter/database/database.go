package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/tiagods/auth/internal/infra/database"
	"github.com/tiagods/auth/internal/infra/message"
	"net/http"
	"time"

	"github.com/tiagods/auth/internal/adapter/database/model"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/httperrors"
)

func (r *repository) GetRefreshToken(ctx context.Context, userID int64, refreshToken *string) (*entity.RefreshToken, error) {
	var refresh model.RefreshToken

	query := FindRefreshToken

	expiresAt := time.Now().UTC()
	vars := []any{expiresAt.Format(time.RFC3339)}

	if userID != 0 {
		query += FindRefreshTokenAddUserID
		vars = append(vars, userID)
	}
	if refreshToken != nil {
		query += FindRefreshTokenAddID
		vars = append(vars, *refreshToken)
	}

	err := r.reader.QueryRow(ctx, query, refresh, vars...)
	if err != nil {
		return nil, err
	}
	if refresh.ID == "" {
		msg := message.ErrLoginRequired
		err = errors.New(msg.UserMessage)
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, err)
	}

	return refresh.ToEntity(), nil

}
func (r *repository) UpdateRefreshToken(ctx context.Context, tx *sql.Tx, userId int64, newToken string, expiresAt time.Time) error {
	var refresh model.RefreshToken
	err := r.reader.QueryRow(ctx, FindRefreshTokenByUser, refresh, userId)
	if err != nil {
		return err
	}

	if refresh.ID == "" {
		_, err = r.writer.Exec(ctx, false, tx, DeleteRefreshToken, refresh.ID)
		if err != nil {
			return err
		}
	}
	now := time.Now().UTC()

	_, err = r.writer.Exec(ctx, false, tx, InsertRefreshToken, newToken, userId, now.Format(time.RFC3339), expiresAt.UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) RegisterAccount(ctx context.Context, tx *sql.Tx, user entity.User) error {
	id, err := r.writer.Exec(ctx, true, tx, InsertUser, user.Username, user.Password)
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

func (r *repository) FindByUserAndPassword(ctx context.Context, username string, password string) (*entity.User, error) {
	var user model.User
	err := r.reader.QueryRow(ctx, FindUserByUserNameAndPassword, user, username, password)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		msg := message.ErrUserNotFound
		err = errors.New(msg.UserMessage)
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, err)
	}

	return user.ToEntity(), nil
}
