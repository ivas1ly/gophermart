package app

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/controller"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	"github.com/ivas1ly/gophermart/internal/repository"
	"github.com/ivas1ly/gophermart/internal/service"
)

type UserService interface {
	Register(ctx context.Context, username, password string) (*entity.User, error)
	Login(ctx context.Context, username, password string) (*entity.User, error)
	NewOrder(ctx context.Context, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetCurrentBalance(ctx context.Context, userID string) (*entity.Balance, error)
	NewWithdrawal(ctx context.Context, userID, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type UserRepository interface {
	Create(ctx context.Context, id, username, hash string) (*entity.User, error)
	Find(ctx context.Context, username string) (*entity.User, error)
	NewOrder(ctx context.Context, orderID, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error)
	NewWithdrawal(ctx context.Context, userID, withdrawalID, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type userServiceProvider struct {
	log            *zap.Logger
	userHandler    *controller.Handler
	userRepository UserRepository
	userService    UserService
}

func newUserServiceProvider(log *zap.Logger) *userServiceProvider {
	return &userServiceProvider{
		log: log.With(zap.String("service", "user")),
	}
}

func (s *userServiceProvider) UserRepository(db *postgres.DB) UserRepository {
	if s.userRepository == nil {
		s.userRepository = repository.NewRepository(db, s.log)
	}
	return s.userRepository
}

func (s *userServiceProvider) UserService(ur UserRepository) UserService {
	if s.userService == nil {
		s.userService = service.NewUserService(ur, s.log)
	}
	return s.userService
}

func (s *userServiceProvider) UserHandler(validate *validator.Validate) *controller.Handler {
	if s.userHandler == nil {
		s.userHandler = controller.NewHandler(s.UserService(s.userRepository), validate, s.log)
	}
	return s.userHandler
}
