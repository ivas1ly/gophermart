package repository

import (
	"context"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	repoEntity "github.com/ivas1ly/gophermart/internal/repository/entity"
)

const defaultWorkerEntities = 5

type Repository struct {
	db  *postgres.DB
	log *zap.Logger
}

func NewRepository(db *postgres.DB, log *zap.Logger) *Repository {
	return &Repository{
		db:  db,
		log: log,
	}
}

func (r *Repository) Create(ctx context.Context, id, username, password string) (*entity.User, error) {
	user := &repoEntity.User{}

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

	query := r.db.Builder.
		Insert("users").
		Columns("id, username, password_hash").
		Values(id, username, password).
		Suffix("RETURNING id, username, password_hash, created_at, updated_at, deleted_at")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, sql, args...)

	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.Hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, entity.ErrUsernameUniqueViolation
		}
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return repoEntity.ToUserFromRepo(user), nil
}

func (r *Repository) Find(ctx context.Context, username string) (*entity.User, error) {
	user := &repoEntity.User{}

	query := r.db.Builder.
		Select("id, username, password_hash, created_at, updated_at, deleted_at").
		From("users").
		Where(sq.Eq{
			"username": username,
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.Pool.QueryRow(ctx, sql, args...)

	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.Hash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrUsernameNotFound
	}
	if err != nil {
		return nil, err
	}

	return repoEntity.ToUserFromRepo(user), nil
}

func (r *Repository) NewOrder(ctx context.Context, orderID, userID, number string) (*entity.Order, error) {
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

	// how to optimize and simplify this check?
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

	r.log.Info("no rows found, continue to add new order")

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

func (r *Repository) GetOrders(ctx context.Context, userID string) ([]entity.Order, error) {
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

func (r *Repository) GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error) {
	userBalance := &repoEntity.Balance{}

	query := r.db.Builder.
		Select("users.id, users.current_balance, COALESCE(SUM(withdrawals.withdrawn), 0)").
		From("users").
		LeftJoin("withdrawals ON withdrawals.user_id = users.id").
		Where(sq.Eq{
			"users.id": userID,
		}).
		GroupBy("users.id")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	row := r.db.Pool.QueryRow(ctx, sql, args...)

	err = row.Scan(
		&userBalance.ID,
		&userBalance.Balance,
		&userBalance.Withdrawn,
	)
	if err != nil {
		return nil, err
	}

	return repoEntity.ToUserBalanceFromRepo(userBalance), nil
}

func (r *Repository) NewWithdrawal(ctx context.Context, userID, withdrawalID, orderNumber string,
	sum int64) error {
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

	queryUpdateBalance := r.db.Builder.
		Update("users").
		SetMap(sq.Eq{
			"current_balance": sq.Expr("current_balance - ?", sum),
		}).
		Where(sq.Eq{
			"id": userID,
		})

	sql, args, err := queryUpdateBalance.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, args...)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			return entity.ErrNotEnoughPointsToWithdraw
		}
		return err
	}

	queryNewWithdrawal := r.db.Builder.
		Insert("withdrawals").
		Columns("id, user_id, order_number, withdrawn").
		Values(withdrawalID, userID, orderNumber, sum)

	sql, args, err = queryNewWithdrawal.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error) {
	query := r.db.Builder.
		Select("id, user_id, order_number, withdrawn, created_at, updated_at, deleted_at").
		From("withdrawals").
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

	withdrawals := make([]repoEntity.Withdraw, 0, DefaultEntityCap)

	for rows.Next() {
		withdraw := repoEntity.Withdraw{}

		err = rows.Scan(
			&withdraw.ID,
			&withdraw.UserID,
			&withdraw.OrderNumber,
			&withdraw.Withdrawn,
			&withdraw.CreatedAt,
			&withdraw.UpdatedAt,
			&withdraw.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdraw)
	}
	if len(withdrawals) == 0 {
		return nil, entity.ErrNoWithdrawalsFound
	}

	return repoEntity.ToWithdrawalsFromRepo(withdrawals), nil
}

func (r *Repository) GetOrdersToProcess(ctx context.Context) ([]entity.Order, error) {
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

func (r *Repository) UpdateOrderAndUserBalance(ctx context.Context, order entity.Order) error {
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
