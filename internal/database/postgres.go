package database

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

func New(ctx context.Context, connString string, logger *zap.Logger) (*pgxpool.Pool, error) {
	var err error
	once.Do(func() {
		config, cfgErr := pgxpool.ParseConfig(connString)
		if cfgErr != nil {
			err = cfgErr
			return
		}

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			return
		}

		if pingErr := pool.Ping(ctx); pingErr != nil {
			err = pingErr
			return
		}
	})

	return pool, err
}

func Get() *pgxpool.Pool {
	return pool
}

func Close() {
	if pool != nil {
		pool.Close()
	}
}
