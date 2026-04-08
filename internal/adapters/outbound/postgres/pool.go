package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool for the hexagonal adapter; use Pgx() for raw access in the same process.
type Pool struct {
	inner *pgxpool.Pool
}

// NewPool opens a PostgreSQL pool. close must be called on shutdown.
func NewPool(ctx context.Context, databaseURL string) (*Pool, func(), error) {
	if databaseURL == "" {
		return nil, nil, fmt.Errorf("database URL is empty")
	}
	pgxPool, connectErr := pgxpool.New(ctx, databaseURL)
	if connectErr != nil {
		return nil, nil, connectErr
	}
	return &Pool{inner: pgxPool}, func() { pgxPool.Close() }, nil
}

// Pgx returns the underlying pool for migrations and repositories.
func (wrapper *Pool) Pgx() *pgxpool.Pool {
	if wrapper == nil {
		return nil
	}
	return wrapper.inner
}

// Ping checks database connectivity.
func (wrapper *Pool) Ping(ctx context.Context) error {
	if wrapper == nil || wrapper.inner == nil {
		return fmt.Errorf("pool not initialized")
	}
	return wrapper.inner.Ping(ctx)
}
