package cache

import (
	"github.com/labstack/gommon/log"
	gocache "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

type cache struct {
	mu    sync.Mutex
	cache *gocache.Cache
}

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

func (c *cache) Get(key string, i interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	i, found := c.cache.Get(key)
	if found {
		return nil
	}
	log.Warnf("cache not found: %s", key)
	return nil
}
