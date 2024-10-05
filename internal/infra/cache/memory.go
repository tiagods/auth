package cache

import (
	"errors"
	"sync"
	"time"

	"github.com/labstack/gommon/log"
	gocache "github.com/patrickmn/go-cache"
)

type cache struct {
	mu    sync.Mutex
	cache *gocache.Cache
}

var (
	ErrNotFound   = errors.New("error not found")
	ErrDuplicated = errors.New("error duplicated key")
)

func NewMemoryCache() Repository {
	c := gocache.New(5*time.Minute, 10*time.Minute)
	return &cache{
		cache: c,
	}
}

func (c *cache) Set(key string, i interface{}, duration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Set(key, i, duration)
	return nil
}

func (c *cache) SetNX(key string, i interface{}, duration time.Duration) error {
	_, err := c.Get(key)
	if err != nil {
		c.Set(key, i, duration)
		return nil
	}

	return ErrDuplicated
}

func (c *cache) Get(key string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	result, found := c.cache.Get(key)
	if found {
		return result, nil
	}

	log.Warnf("cache not found: %s", key)
	return nil, ErrNotFound
}
