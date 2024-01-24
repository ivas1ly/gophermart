package service

import (
	"context"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type UserService interface {
	Register(ctx context.Context, username, password string) (*entity.User, error)
	Login(ctx context.Context, username, password string) (*entity.User, error)
}
