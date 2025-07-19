package cache

import (
	"context"
	"time"
)

type Repository interface {
	IsAlive() bool
	Set(ctx context.Context, key string, i interface{}, duration time.Duration) error
	SetNX(ctx context.Context, key string, i interface{}, duration time.Duration) error
	Get(ctx context.Context, key string, output interface{}) error
	Delete(ctx context.Context, key string) error
}

func NewCache() Repository {
	return NewRedis()
}
