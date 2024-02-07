package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/utils/argon2id"
)

type Repository interface {
	Create(ctx context.Context, id, username, hash string) (*entity.User, error)
	Find(ctx context.Context, username string) (*entity.User, error)
	NewOrder(ctx context.Context, orderID, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
	GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error)
	NewWithdrawal(ctx context.Context, userID, withdrawalID, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type UserService struct {
	userRepository Repository
	log            *zap.Logger
}

func NewUserService(userRepository Repository, log *zap.Logger) *UserService {
	return &UserService{
		userRepository: userRepository,
		log:            log,
	}
}

func (s *UserService) Register(ctx context.Context, username, password string) (*entity.User, error) {
	userUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.Create(ctx, userUUID.String(), username, hash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.userRepository.Find(ctx, username)
	if err != nil {
		return nil, err
	}

	ok, err := argon2id.ComparePasswordAndHash(password, user.Hash)
	if !ok {
		return nil, entity.ErrIncorrectLoginOrPassword
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) NewOrder(ctx context.Context, userID, orderNumber string) (*entity.Order, error) {
	orderUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	order, err := s.userRepository.NewOrder(ctx, orderUUID.String(), userID, orderNumber)
	if errors.Is(err, entity.ErrOrderUniqueViolation) {
		if order.UserID == userID {
			return nil, entity.ErrUploadedByThisUser
		}
		return nil, entity.ErrUploadedByAnotherUser
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *UserService) GetOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	orders, err := s.userRepository.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *UserService) GetCurrentBalance(ctx context.Context, userID string) (*entity.Balance, error) {
	currentBalance, err := s.userRepository.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return currentBalance, nil
}

func (s *UserService) NewWithdrawal(ctx context.Context, userID, orderNumber string, sum int64) error {
	withdrawalUUID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	err = s.userRepository.NewWithdrawal(ctx, userID, withdrawalUUID.String(), orderNumber, sum)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error) {
	withdrawals, err := s.userRepository.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
