package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/tiagods/auth/internal/infra/database"
	"github.com/tiagods/auth/internal/infra/logger"
	"time"

	"github.com/tiagods/auth/internal/infra/tracer"

	"github.com/tiagods/auth/internal/adapter/database/model"
	"github.com/tiagods/auth/internal/domain/entity"
)

func (r *repository) GetRefreshToken(ctx context.Context, userID int64, refreshToken *string) (*entity.RefreshToken, error) {
	caller := "repository::get_refresh_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	refresh := model.RefreshToken{}

	query := FindRefreshToken

	expiresAt := time.Now()
	vars := []any{expiresAt}

	if userID != 0 {
		query += FindRefreshTokenAddUserID
		vars = append(vars, userID)
	}
	if refreshToken != nil {
		query += FindRefreshTokenAddID
		vars = append(vars, *refreshToken)
	}

	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &refresh, query, vars...)
	if err != nil {
		return nil, err
	}
	if refresh.ID == "" {
		return nil, ErrNotFound
	}

	return refresh.ToEntity(), nil

}

func (r *repository) DeleteRefreshToken(ctx context.Context, tx *sql.Tx, userId int64) error {
	caller := "repository::delete_refresh_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	refresh := model.RefreshToken{}
	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &refresh, FindRefreshTokenByUser, userId)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}

	if refresh.ID == "" {
		return ErrNotFound
	}

	_, err = r.writer.Exec(ctx, false, tx, DeleteRefreshToken, refresh.ID)
	if err != nil {
		if errors.Is(err, database.ErrNoRowsAffected) {
			tracer.SetError(span, err)
			logger.Warn(ctx, err, "no rows affected when trying to delete refresh token")
			return nil
		}
		tracer.SetError(span, err)
		logger.Error(ctx, err, "failed to delete refresh token")
		return err
	}

	return nil
}

func (r *repository) UpdateRefreshToken(ctx context.Context, tx *sql.Tx, userId int64, newToken string, expiresAt time.Time) error {
	caller := "repository::update_refresh_token"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	refresh := model.RefreshToken{}
	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &refresh, FindRefreshTokenByUser, userId)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}

	_, err = r.writer.Exec(ctx, false, tx, DeleteRefreshToken, refresh.ID)
	if err != nil {
		if errors.Is(err, database.ErrNoRowsAffected) {
			tracer.SetError(span, err)
			logger.Warn(ctx, err, "no rows affected when trying to delete old refresh token")
		} else {
			tracer.SetError(span, err)
			logger.Error(ctx, err, "failed to delete old refresh token")
			return err
		}
	}

	now := time.Now()
	_, err = r.writer.Exec(ctx, false, tx, InsertRefreshToken, newToken, userId, now, expiresAt)
	if err != nil {
		tracer.SetError(span, err)
		return err
	}

	return nil
}

func (r *repository) RegisterAccount(ctx context.Context, tx *sql.Tx, user *entity.UserCredential) error {
	caller := "repository::register_account"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	id, err := r.writer.Exec(ctx, true, tx, InsertUser, user.Username, user.Password)
	if err != nil {
		tracer.SetError(span, err)
		logger.Error(ctx, err, "failed to register user account")
		return err
	}
	user.ID = id
	return nil
}

func (r *repository) FindByUserAndPassword(ctx context.Context, username string, password string) (*entity.UserCredential, error) {
	caller := "repository::find_by_user_and_password"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	user := model.User{}
	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &user, FindUserByUserNameAndPassword, username, password)
	if err != nil {
		tracer.SetError(span, err)
		return nil, err
	}
	if user.ID == 0 {
		return nil, ErrNotFound
	}

	return user.ToEntity(), nil
}

func (r *repository) ListUsers(ctx context.Context, offset int, limit int) ([]*entity.UserCredential, bool, error) {
	caller := "repository::list_users"
	ctx, span := tracer.Start(ctx, caller, tracer.SpanKindCPU)
	defer span.End()

	if offset > 0 {
		offset = offset - 1
	}
	newLimit := limit + 1
	rows, finish, err := r.reader.QueryRows(ctx, ListUsers, newLimit, offset).Result()
	defer finish()
	if err != nil {
		return nil, false, err
	}

	var users []*entity.UserCredential
	for rows.Next() {
		user := model.User{}
		err = rows.Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			return nil, false, err
		}
		users = append(users, user.ToEntity())
	}
	hasNextPage := len(users) > limit
	if hasNextPage {
		users = users[:len(users)-1]
	}
	return users, hasNextPage, err
}
