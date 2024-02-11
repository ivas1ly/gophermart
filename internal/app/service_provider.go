package app

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	auth "github.com/ivas1ly/gophermart/internal/api/controller/auth"
	balance "github.com/ivas1ly/gophermart/internal/api/controller/balance"
	order "github.com/ivas1ly/gophermart/internal/api/controller/order"
	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	"github.com/ivas1ly/gophermart/internal/repository"
	"github.com/ivas1ly/gophermart/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*entity.User, error)
	Login(ctx context.Context, username, password string) (*entity.User, error)
}

type OrderService interface {
	NewOrder(ctx context.Context, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type BalanceService interface {
	GetCurrentBalance(ctx context.Context, userID string) (*entity.Balance, error)
	NewWithdrawal(ctx context.Context, userID, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type AccrualWorkerService interface {
	GetNewOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrders(ctx context.Context, orders ...entity.Order) error
}

type AuthRepository interface {
	Create(ctx context.Context, id, username, hash string) (*entity.User, error)
	Find(ctx context.Context, username string) (*entity.User, error)
}

type OrderRepository interface {
	NewOrder(ctx context.Context, orderID, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error)
	NewWithdrawal(ctx context.Context, userID, withdrawalID, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type AccrualWorkerRepository interface {
	GetOrdersToProcess(ctx context.Context) ([]entity.Order, error)
	UpdateOrderAndUserBalance(ctx context.Context, order entity.Order) error
}

type ServiceProvider struct {
	OrderService         OrderService
	AuthService          AuthService
	BalanceService       BalanceService
	AccrualWorkerService AccrualWorkerService

	db *postgres.DB
}

func NewServiceProvider(db *postgres.DB) *ServiceProvider {
	return &ServiceProvider{
		db: db,
	}
}

func (s *ServiceProvider) RegisterServices() {
	s.NewOrderService()
	s.NewAuthService()
	s.NewBalanceService()
}

func (s *ServiceProvider) RegisterHandlers(router *chi.Mux, validate *validator.Validate) *chi.Mux {
	router.Route("/api/user", func(r chi.Router) {
		auth.NewAuthHandler(s.AuthService, validate).Register(r)
		order.NewOrderHandler(s.OrderService).Register(r)
		balance.NewBalanceHandler(s.BalanceService, validate).Register(r)
	})

	return router
}

func (s *ServiceProvider) newAuthRepository() AuthRepository {
	return repository.NewAuthRepository(s.db)
}

func (s *ServiceProvider) NewAuthService() AuthService {
	if s.AuthService == nil {
		s.AuthService = service.NewAuthService(s.newAuthRepository())
	}

	return s.AuthService
}

func (s *ServiceProvider) newOrderRepository() OrderRepository {
	return repository.NewOrderRepository(s.db)
}

func (s *ServiceProvider) NewOrderService() OrderService {
	if s.OrderService == nil {
		s.OrderService = service.NewOrderService(s.newOrderRepository())
	}

	return s.OrderService
}

func (s *ServiceProvider) newBalanceRepository() BalanceRepository {
	return repository.NewBalanceRepository(s.db)
}

func (s *ServiceProvider) NewBalanceService() BalanceService {
	if s.BalanceService == nil {
		s.BalanceService = service.NewBalanceService(s.newBalanceRepository())
	}

	return s.BalanceService
}
