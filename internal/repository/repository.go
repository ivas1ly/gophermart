package repository

import (
	"context"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type UserRepository interface {
	Create(ctx context.Context, id, username, hash string) (*entity.User, error)
	Find(ctx context.Context, username string) (*entity.User, error)
}
