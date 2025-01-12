package database_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tiagods/auth/internal/adapter/database"
	databaseType "github.com/tiagods/auth/internal/infra/database"
	"github.com/tiagods/auth/internal/infra/requestcontext"
	"testing"
)

func TestRepository_FindByUserAndPassword(t *testing.T) {
	db := databaseType.NewDB("root", "auth", "localhost:3306", "auth")
	repo := database.NewRepository(db, db)
	ctx := context.WithValue(context.Background(), requestcontext.ContextKey, requestcontext.RequestContext{
		Cid:    "cid-test",
		Tenant: "tenant-test",
		Roles:  nil,
	})

	t.Run("FindByUserAndPassword", func(t *testing.T) {
		user, err := repo.FindByUserAndPassword(ctx, "unknow", "password")
		assert.Error(t, err)
		assert.NotNil(t, user)
	})

	t.Run("FindByUserAndPassword", func(t *testing.T) {
		user, err := repo.FindByUserAndPassword(ctx, "tiago", "tiago")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.True(t, user.ID != 0)
	})
}
