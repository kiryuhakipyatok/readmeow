package cache

import (
	"context"
	"readmeow/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Redis *redis.Client
}

const Empty = redis.Nil

func Connect(cfg config.CacheConfig) Cache {
	redis := redis.NewClient(&redis.Options{
		Username: cfg.User,
		Addr:     cfg.Host + ":" + cfg.Port,
		DB:       0,
		Password: cfg.Password,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := redis.Ping(ctx).Err(); err != nil {
		panic("failed to ping redis client")
	}
	cache := Cache{
		Redis: redis,
	}
	return cache
}

func (c *Cache) Close() {
	if err := c.Redis.Close(); err != nil {
		panic("failed to close redis")
	}
}
