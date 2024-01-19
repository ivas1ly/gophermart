package user

import (
	"context"

	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	def "github.com/ivas1ly/gophermart/internal/repository"
)

var _ def.UserRepository = (*Repository)(nil)

type Repository struct {
	db *postgres.DB
}

func NewRepository(db *postgres.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context) error {
	return nil
}
