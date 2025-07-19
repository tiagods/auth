package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/gommon/log"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func (r *redisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
	}
	return nil
}

func NewRedis() Repository {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return &redisCache{
		client: rdb,
	}
}

func (r *redisCache) IsAlive() bool {
	if err := r.client.Ping(context.Background()).Err(); err != nil {
		log.Error(err.Error())
		return false
	}
	return true
}

func (r *redisCache) Set(ctx context.Context, key string, i interface{}, duration time.Duration) error {
	data, err := json.Marshal(i)
	if err != nil {
		return err
	}
	err = r.client.Set(ctx, key, data, duration).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *redisCache) SetNX(ctx context.Context, key string, i interface{}, duration time.Duration) error {
	err := r.client.SetNX(ctx, key, i, duration).Err()
	if err != nil {
		return err
	}
	return nil
}
func (r *redisCache) Get(ctx context.Context, key string, output interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrNotFound
		}
		return err
	}
	err = json.Unmarshal([]byte(val), output)
	if err != nil {
		return err
	}

	return err
}
