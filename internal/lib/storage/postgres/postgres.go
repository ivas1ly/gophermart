package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	Pool    *pgxpool.Pool
	Builder squirrel.StatementBuilderType
}

func New(ctx context.Context, dsn string, attempts int, timeout time.Duration) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("can't create new connection pool: %w", err)
	}

	err = retryWithAttempts(func() error {
		if err = pool.Ping(ctx); err != nil {
			return fmt.Errorf("failed to ping database: %w", err)
		}
		return nil
	}, attempts, timeout)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &DB{
		Pool:    nil,
		Builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar), // PostgreSQL placeholder format
	}, err
}

func retryWithAttempts(fn func() error, attempts int, timeout time.Duration) error {
	var err error

	for attempts > 0 {
		if err = fn(); err != nil {
			zap.L().Info("trying to connect Postgres...", zap.Int("attempts left", attempts))
			time.Sleep(timeout)
			attempts--
			continue
		}
		return nil
	}

	return err
}
