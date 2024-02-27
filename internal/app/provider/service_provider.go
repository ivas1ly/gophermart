package provider

import (
	"context"

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
	AddOrder(ctx context.Context, orderInfo *entity.OrderInfo) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type BalanceService interface {
	GetCurrentBalance(ctx context.Context, userID string) (*entity.Balance, error)
	AddWithdrawal(ctx context.Context, withdrawInfo *entity.WithdrawInfo) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type AccrualWorkerService interface {
	GetNewOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrders(ctx context.Context, orders ...entity.Order) error
}

type AuthRepository interface {
	AddUser(ctx context.Context, userInfo *entity.UserInfo) (*entity.User, error)
	FindUser(ctx context.Context, username string) (*entity.User, error)
}

type OrderRepository interface {
	AddOrder(ctx context.Context, orderInfo *entity.OrderInfo) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error)
	AddWithdrawal(ctx context.Context, withdrawInfo *entity.WithdrawInfo) error
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
