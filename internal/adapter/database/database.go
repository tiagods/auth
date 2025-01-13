package database

import (
	"context"
	"database/sql"
	"github.com/tiagods/auth/internal/infra/message"
	"github.com/tiagods/auth/internal/infra/otel"
	"net/http"
	"time"

	"github.com/tiagods/auth/internal/adapter/database/model"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/httperrors"
)

func (r *repository) GetRefreshToken(ctx context.Context, userID int64, refreshToken *string) (*entity.RefreshToken, error) {
	caller := "repository::get_refresh_token"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	refresh := model.RefreshToken{}

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

	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &refresh, query, vars...)
	if err != nil {
		return nil, err
	}
	if refresh.ID == "" {
		msg := message.ErrLoginRequired
		otel.SetError(span, err)
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	return refresh.ToEntity(), nil

}
func (r *repository) UpdateRefreshToken(ctx context.Context, tx *sql.Tx, userId int64, newToken string, expiresAt time.Time) error {
	caller := "repository::update_refresh_token"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	refresh := model.RefreshToken{}
	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &refresh, FindRefreshTokenByUser, userId)
	if err != nil {
		otel.SetError(span, err)
		return err
	}

	if refresh.ID == "" {
		_, err = r.writer.Exec(ctx, false, tx, DeleteRefreshToken, refresh.ID)
		if err != nil {
			otel.SetError(span, err)
			return err
		}
	}
	now := time.Now().UTC()

	_, err = r.writer.Exec(ctx, false, tx, InsertRefreshToken, newToken, userId, now.Format(time.RFC3339), expiresAt.UTC().Format(time.RFC3339))
	if err != nil {
		otel.SetError(span, err)
		return err
	}

	return nil
}

func (r *repository) RegisterAccount(ctx context.Context, tx *sql.Tx, user entity.User) error {
	caller := "repository::register_account"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	id, err := r.writer.Exec(ctx, true, tx, InsertUser, user.Username, user.Password)
	if err != nil {
		otel.SetError(span, err)
		return err
	}
	user.ID = id
	return nil
}

func (r *repository) FindByUserAndPassword(ctx context.Context, username string, password string) (*entity.User, error) {
	caller := "repository::find_by_user_and_password"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
	defer span.End()

	user := model.User{}
	err := r.reader.GetSqlScanner(ctx).Select(r.reader.DB, &user, FindUserByUserNameAndPassword, username, password)
	if err != nil {
		otel.SetError(span, err)
		return nil, err
	}
	if user.ID == 0 {
		msg := message.ErrUserNotFound
		otel.SetError(span, msg.GetError())
		return nil, httperrors.NewHttpError(ctx, http.StatusUnauthorized, msg, msg.GetError())
	}

	return user.ToEntity(), nil
}

func (r *repository) ListUsers(ctx context.Context, offset int, limit int) ([]*entity.User, bool, error) {
	caller := "repository::list_users"
	ctx, span := otel.Start(ctx, caller, otel.SpanKingCPU)
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

	var users []*entity.User
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
