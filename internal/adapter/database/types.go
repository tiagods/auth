package database

import (
	"context"

	"github.com/tiagods/auth/internal/adapter/database/model"
)

type (
	Repository interface {
		//RegisterAccount(ctx context.Context, user entity.User) error
		FindRefreshToken(ctx context.Context, refreshToken string) (model.User, error)
		UpdateRefreshToken(ctx context.Context, userId string, newToken string) error
		FindByUserAndPassword(ctx context.Context, username string, password string) (model.User, error)
	}
)
