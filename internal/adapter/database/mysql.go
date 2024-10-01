package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/tiagods/auth/internal/adapter/database/entity"
	"github.com/tiagods/auth/internal/infra/httperrors"
	"net/http"
)

type (
	databaseRepository struct {
		db *sql.DB
	}
)

func NewDatabaseRepository(db *sql.DB) Repository {
	return databaseRepository{db: db}
}

func (m databaseRepository) RegisterAccount(ctx context.Context, user entity.User) error {
	return nil
}

func (m databaseRepository) FindByUserAndPassword(ctx context.Context, username string, password string) (entity.User, error) {
	var user entity.User
	query := `SELECT username, password FROM Users WHERE username=? AND password=?`
	stmt, err := m.db.PrepareContext(ctx, query)
	row := stmt.QueryRow(username, password)

	err = row.Scan(&user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = errors.New("no account found")
			return entity.User{}, httperrors.NewHttpError(http.StatusUnauthorized, err.Error(), err)
		} else {
			return entity.User{}, err
		}
	}
	return user, nil
}
