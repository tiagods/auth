package model

import (
	"github.com/tiagods/auth/internal/domain/entity"
	"time"
)

type (
	RefreshToken struct {
		ID        string    `db:"id"`
		UserID    int64     `db:"user_id"`
		CreatedAt time.Time `db:"created_at"`
		ExpiresAt time.Time `db:"expires_at"`
	}
)

func (t RefreshToken) ToEntity() *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:     t.ID,
		UserID: t.UserID,
	}
}
