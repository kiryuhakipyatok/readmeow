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

func MustConnect(cfg config.StorageConfig) *Storage {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&timezone=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
		cfg.Timezone,
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.ConnectTimeout))
	defer cancel()
	pcfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic("failed to parse pool config" + err.Error())
	}
	pcfg.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		panic("failed to create postgres pool" + err.Error())
	}
	if err := pool.Ping(ctx); err != nil {
		panic("failed to ping postgres pool" + err.Error())
	}
	storage := &Storage{
		Pool: pool,
	}
	return storage
}

func (s *Storage) Close() {
	s.Pool.Close()
}
