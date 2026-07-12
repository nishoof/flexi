package database

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nishoof/flexi/backend/repository"
)

var (
	pool    *pgxpool.Pool
	once    sync.Once
	initErr error
)

// Returns a lazily-initialized pgx pool, reused across warm Lambda invocations.
func getPool(ctx context.Context) (*pgxpool.Pool, error) {
	once.Do(func() {
		connString := os.Getenv("DATABASE_URL")
		if connString == "" {
			initErr = errors.New("DATABASE_URL is not set")
			return
		}
		config, err := pgxpool.ParseConfig(connString)
		if err != nil {
			initErr = err
			return
		}

		// 1 connection to Supavisor (pooler, like pgbouncer)
		config.MaxConns = 1

		// Disable caching stuff such as prepared statements
		// Transaction pooling can give a different connection, so caching causes issues
		// For example, if we prepare a statement on one connection, then the next transaction is on a different connection, the prepared statement won't exist there.
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
		config.ConnConfig.StatementCacheCapacity = 0
		config.ConnConfig.DescriptionCacheCapacity = 0

		pool, initErr = pgxpool.NewWithConfig(ctx, config)
	})
	return pool, initErr
}

func Queries(ctx context.Context) (*repository.Queries, error) {
	pool, err := getPool(ctx)
	if err != nil {
		return nil, err
	}
	return repository.New(pool), nil
}

func QueriesWithTx(ctx context.Context) (*repository.Queries, pgx.Tx, error) {
	pool, err := getPool(ctx)
	if err != nil {
		return nil, nil, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	qtx := repository.New(tx)

	return qtx, tx, nil
}
