package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/labstack/gommon/log"
	gocache "github.com/patrickmn/go-cache"
)

type memory struct {
	mu    sync.Mutex
	cache *gocache.Cache
}

var (
	ErrNotFound   = errors.New("error not found")
	ErrDuplicated = errors.New("error duplicated key")
)

func NewMemoryCache() Repository {
	c := gocache.New(5*time.Minute, 10*time.Minute)
	return &memory{
		cache: c,
	}
}

func (c *memory) IsAlive() bool {
	return true
}

func (c *memory) Set(ctx context.Context, key string, i interface{}, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Set(key, i, duration)
	return nil
}

func (c *memory) SetNX(ctx context.Context, key string, i interface{}, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.cache.Get(key)
	if found {
		return ErrDuplicated
	}
	return c.Set(ctx, key, i, duration)

}

func (c *memory) Get(ctx context.Context, key string, output interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	output, found := c.cache.Get(key)
	if found {
		return nil
	}

	log.Warnf("%s : %s", ErrNotFound.Error(), key)
	return ErrNotFound
}
