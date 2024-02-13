package repository

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	repoEntity "github.com/ivas1ly/gophermart/internal/repository/entity"
)

type BalanceRepository struct {
	db *postgres.DB
}

func NewBalanceRepository(db *postgres.DB) *BalanceRepository {
	return &BalanceRepository{
		db: db,
	}
}

func (r *BalanceRepository) GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error) {
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

func (r *BalanceRepository) AddWithdrawal(ctx context.Context, userID, withdrawalID, orderNumber string,
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

func (r *BalanceRepository) GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error) {
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
