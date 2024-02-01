package repository

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	repoEntity "github.com/ivas1ly/gophermart/internal/repository/entity"
)

type UserRepository interface {
	Create(ctx context.Context, id, username, hash string) (*entity.User, error)
	Find(ctx context.Context, username string) (*entity.User, error)
	NewOrder(ctx context.Context, orderID, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetUserBalance(ctx context.Context, userID string) (*entity.UserBalance, error)
}

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

func (r *Repository) GetUserBalance(ctx context.Context, userID string) (*entity.UserBalance, error) {
	userBalance := &repoEntity.UserBalance{}

	query := r.db.Builder.
		Select("users.id, users.current_balance, SUM(withdrawals.withdrawn)").
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
