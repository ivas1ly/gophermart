package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID string) (*entity.Balance, error)
	NewWithdrawal(ctx context.Context, userID, withdrawalID, orderNumber string, sum int64) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type BalanceService struct {
	balanceRepository BalanceRepository
}

func NewBalanceService(balanceRepository BalanceRepository) *BalanceService {
	return &BalanceService{
		balanceRepository: balanceRepository,
	}
}

func (s *BalanceService) GetCurrentBalance(ctx context.Context, userID string) (*entity.Balance, error) {
	currentBalance, err := s.balanceRepository.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	return currentBalance, nil
}

func (s *BalanceService) NewWithdrawal(ctx context.Context, userID, orderNumber string, sum int64) error {
	withdrawalUUID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	err = s.balanceRepository.NewWithdrawal(ctx, userID, withdrawalUUID.String(), orderNumber, sum)
	if err != nil {
		return err
	}

	return nil
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error) {
	withdrawals, err := s.balanceRepository.GetWithdrawals(ctx, userID)
	if err != nil {
		return nil, err
	}

	return withdrawals, nil
}
