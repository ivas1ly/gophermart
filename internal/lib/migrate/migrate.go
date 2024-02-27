package migrate

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/migrations"
)

func Run(ctx context.Context, pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		&migrations.Migrations,
	)
	if err != nil {
		return fmt.Errorf("can't create new goose provider: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("can't up migrations: %w", err)
	}

	if len(results) == 0 {
		zap.L().Info("no change to database schema")
	}

	err = db.Close()
	if err != nil {
		return fmt.Errorf("can't close the database: %w", err)
	}

	return nil
}
