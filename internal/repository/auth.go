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

const defaultWorkerEntities = 5

type AuthRepository struct {
	db *postgres.DB
}

func NewAuthRepository(db *postgres.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) AddUser(ctx context.Context, userInfo *entity.UserInfo) (*entity.User, error) {
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
		Values(userInfo.ID, userInfo.Username, userInfo.Hash).
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

func (r *AuthRepository) FindUser(ctx context.Context, username string) (*entity.User, error) {
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
