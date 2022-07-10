package db

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache interface {
	Add(context.Context, string, time.Duration) error
	Exist(context.Context, string) (bool, error)
	Delete(context.Context, string) error
}

type RedisCache struct {
	*redis.Client
}

func NewRedis(ctx context.Context, host string, port int) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &RedisCache{rdb}, err
}

func (c *RedisCache) Add(ctx context.Context, key string, expiration time.Duration) error {
	return c.Set(ctx, key, "value", expiration*time.Second).Err()
}

func (c *RedisCache) Exist(ctx context.Context, key string) (bool, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val != "", nil
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	if _, err := c.Del(ctx, key).Result(); err != nil {
		return err
	}

	return nil
}
