package cache

import "time"

type Repository interface {
	Set(key string, i interface{}, duration time.Duration) error
	SetNX(key string, i interface{}, duration time.Duration) error
	Get(key string) (interface{}, error)
}
