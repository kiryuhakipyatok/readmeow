package storage

import (
	"context"
	"fmt"
	"readmeow/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Pool *pgxpool.Pool
}

func Connect(cfg *config.StorageConfig) *Storage {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
		cfg.Timezone,
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic("failed to create postgres pool")
	}
	if err := pool.Ping(ctx); err != nil {
		panic("failed to ping postgres pool")
	}
	storage := &Storage{
		Pool: pool,
	}
	return storage
}

func (s *Storage) Close() {
	s.Pool.Close()
}
