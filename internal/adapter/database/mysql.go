package database

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/tiagods/auth/internal/adapter/database/model"
	"github.com/tiagods/auth/internal/domain/entity"
	"github.com/tiagods/auth/internal/infra/httperrors"
)

func (r *repository) FindRefreshToken(ctx context.Context, refreshToken string) (model.User, error) {
	return model.User{}, nil

}
func (r *repository) UpdateRefreshToken(ctx context.Context, userId string, newToken string) error {
	return nil
}

func (r *repository) RegisterAccount(ctx context.Context, user entity.User) error {
	return nil
}

func (r *repository) FindByUserAndPassword(ctx context.Context, username string, password string) (model.User, error) {
	var user model.User
	query := `SELECT username, password FROM Users WHERE username=? AND password=?`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return model.User{}, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(username, password)

	err = row.Scan(&user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("no account found")
			return model.User{}, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
		} else {
			return model.User{}, err
		}
	}
	return user, nil
}
