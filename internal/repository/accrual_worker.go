package repository

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	repoEntity "github.com/ivas1ly/gophermart/internal/repository/entity"
)

type AccrualWorkerRepository struct {
	db *postgres.DB
}

func NewAccrualWorkerRepository(db *postgres.DB) *AccrualWorkerRepository {
	return &AccrualWorkerRepository{
		db: db,
	}
}

func (r *AccrualWorkerRepository) GetOrdersToProcess(ctx context.Context) ([]entity.Order, error) {
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

	querySelect := r.db.Builder.
		Select("id, user_id, number, status, accrual, created_at, updated_at, deleted_at").
		From("orders").
		Where(sq.Or{
			sq.Eq{
				"status": entity.StatusNew.String(),
			},
			sq.Eq{
				"status": entity.StatusProcessing.String(),
			},
		}).
		OrderBy("created_at ASC").
		Limit(defaultWorkerEntities)

	queryUpdate := r.db.Builder.
		Update("orders").
		SetMap(
			sq.Eq{
				"status":     entity.StatusProcessing.String(),
				"updated_at": time.Now(),
			},
		).
		FromSelect(querySelect, "subquery").
		Suffix("RETURNING orders.id, orders.user_id, orders.number, orders.status, " +
			"orders.accrual, orders.created_at, orders.updated_at, orders.deleted_at")

	sql, args, err := queryUpdate.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return repoEntity.ToOrdersFromRepo(orders), nil
}

func (r *AccrualWorkerRepository) UpdateOrderAndUserBalance(ctx context.Context, order entity.Order) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func(tx pgx.Tx) {
		err = tx.Rollback(ctx)
		if errors.Is(err, pgx.ErrTxClosed) {
			return
		}
	}(tx)

	queryUpdateOrders := r.db.Builder.Update("orders").
		SetMap(sq.Eq{
			"accrual": order.Accrual,
			"status":  order.Status,
		}).
		Where(sq.Eq{
			"id": order.ID,
		}).
		Suffix("RETURNING id, user_id, number, status, accrual, created_at, updated_at, deleted_at")

	sqlUpdate, args, err := queryUpdateOrders.ToSql()
	if err != nil {
		return err
	}

	rowUpdate := tx.QueryRow(ctx, sqlUpdate, args...)

	var updateOrderResult entity.Order

	err = rowUpdate.Scan(
		&updateOrderResult.ID,
		&updateOrderResult.UserID,
		&updateOrderResult.Number,
		&updateOrderResult.Status,
		&updateOrderResult.Accrual,
		&updateOrderResult.CreatedAt,
		&updateOrderResult.UpdatedAt,
		&updateOrderResult.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err != nil {
		return errors.Join(err, entity.ErrCanNotUpdateOrder)
	}

	queryUpdateBalance := r.db.Builder.
		Update("users").
		SetMap(sq.Eq{
			"current_balance": sq.Expr("current_balance + ?", order.Accrual),
		}).
		Where(sq.Eq{
			"id": updateOrderResult.UserID,
		}).
		Suffix("RETURNING id, current_balance, username")

	sqlUpdate, args, err = queryUpdateBalance.ToSql()
	if err != nil {
		return err
	}

	rowUpdate = tx.QueryRow(ctx, sqlUpdate, args...)

	var updateUserBalanceResult entity.User

	err = rowUpdate.Scan(
		&updateUserBalanceResult.ID,
		&updateUserBalanceResult.Balance,
		&updateUserBalanceResult.Username,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err != nil {
		return errors.Join(err, entity.ErrCanNotUpdateUserBalance)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
