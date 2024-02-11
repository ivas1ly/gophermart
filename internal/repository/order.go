package repository

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	repoEntity "github.com/ivas1ly/gophermart/internal/repository/entity"
)

type OrderRepository struct {
	db *postgres.DB
}

func NewOrderRepository(db *postgres.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) NewOrder(ctx context.Context, orderID, userID, number string) (*entity.Order, error) {
	order := &repoEntity.Order{}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func(tx pgx.Tx) {
		err = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrTxClosed) {
			return
		}
	}(tx)

	checkQuery := r.db.Builder.
		Select("id, user_id, number, status, accrual, created_at, updated_at, deleted_at").
		From("orders").
		Where(sq.Eq{
			"number": number,
		})

	sql, args, err := checkQuery.ToSql()
	if err != nil {
		return nil, err
	}

	checkRow := tx.QueryRow(ctx, sql, args...)
	err = checkRow.Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.DeletedAt,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return repoEntity.ToOrderFromRepo(order), entity.ErrOrderUniqueViolation
	}

	zap.L().Info("no rows found, continue to add new order")

	query := r.db.Builder.
		Insert("orders").
		Columns("id, user_id, number, status, accrual").
		Values(orderID, userID, number, entity.StatusNew.String(), 0).
		Suffix("RETURNING id, user_id, number, status, accrual, created_at, updated_at, deleted_at")

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, sql, args...)
	err = row.Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return repoEntity.ToOrderFromRepo(order), nil
}

func (r *OrderRepository) GetOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	query := r.db.Builder.
		Select("id, user_id, number, status, accrual, created_at, updated_at, deleted_at").
		From("orders").
		Where(sq.Eq{
			"user_id": userID,
		}).
		OrderBy("created_at ASC").
		Limit(DefaultEntityCap)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]repoEntity.Order, 0, DefaultEntityCap)

	for rows.Next() {
		order := repoEntity.Order{}

		err = rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.CreatedAt,
			&order.UpdatedAt,
			&order.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}
	if len(orders) == 0 {
		return nil, entity.ErrNoOrdersFound
	}

	return repoEntity.ToOrdersFromRepo(orders), nil
}
