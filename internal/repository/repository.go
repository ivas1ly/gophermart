package repository

import "context"

type UserRepository interface {
	Create(ctx context.Context) error
}
