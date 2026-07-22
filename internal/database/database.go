package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	connectTimeout = 5 * time.Second
	maxRetries     = 5
	retryDelay     = 500 * time.Millisecond
)

// Open creates a pgx pool for databaseURL, pinging until ready or retries are exhausted.
func Open(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}

	var pool *pgxpool.Pool
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptCtx, cancel := context.WithTimeout(ctx, connectTimeout)
		pool, err = pgxpool.NewWithConfig(attemptCtx, cfg)
		if err == nil {
			err = pool.Ping(attemptCtx)
		}
		cancel()

		if err == nil {
			return pool, nil
		}
		lastErr = err
		if pool != nil {
			pool.Close()
			pool = nil
		}
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("connect to database: %w", ctx.Err())
			case <-time.After(retryDelay):
			}
		}
	}
	return nil, fmt.Errorf("connect to database after %d attempts: %w", maxRetries, lastErr)
}

// Close closes the pool when non-nil.
func Close(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}
