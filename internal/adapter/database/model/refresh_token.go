package model

import (
	"github.com/tiagods/auth/internal/domain/entity"
	"time"
)

type (
	RefreshToken struct {
		ID        string
		UserID    int64
		CreatedAt time.Time
		ExpiresAt time.Time
	}
)

func (t RefreshToken) ToEntity() *entity.RefreshToken {
	return &entity.RefreshToken{
		ID:     t.ID,
		UserID: t.UserID,
	}
}
