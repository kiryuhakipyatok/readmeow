package cache

import (
	"context"
	"fmt"
	"readmeow/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Redis *redis.Client
}

const EMPTY = redis.Nil

func MustConnect(cfg config.CacheConfig) *Cache {

	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		DB:       0,
		Password: cfg.Password,
		PoolSize: 10,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.ConnectTimeout))
	defer cancel()
	if err := redis.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("failed to ping redis client: %w", err))
	}
	cache := &Cache{
		Redis: redis,
	}
	return cache
}

func (c *Cache) MustClose() {
	if err := c.Redis.Close(); err != nil {
		panic(fmt.Errorf("failed to close redis: %w", err))
	}
}
