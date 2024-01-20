package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	def "github.com/ivas1ly/gophermart/internal/repository"
)

var _ def.UserRepository = (*Repository)(nil)

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

func (r *Repository) Create(ctx context.Context) error {
	return nil
}
