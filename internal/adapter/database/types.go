package database

import (
	"context"
	"github.com/tiagods/auth/internal/adapter/database/entity"
)

type (
	Repository interface {
		//RegisterAccount(ctx context.Context, user entity.User) error
		FindByUserAndPassword(ctx context.Context, username string, password string) (entity.User, error)
	}
)
