package controller

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type BalanceService interface {
	GetCurrentBalance(ctx context.Context, userID string) (*entity.Balance, error)
	AddWithdrawal(ctx context.Context, withdrawInfo *entity.WithdrawInfo) error
	GetWithdrawals(ctx context.Context, userID string) ([]entity.Withdraw, error)
}

type BalanceHandler struct {
	balanceService BalanceService
	log            *zap.Logger
	validate       *validator.Validate
}

func NewBalanceHandler(balanceService BalanceService, validate *validator.Validate) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
		log:            zap.L().With(zap.String("handler", "balance")),
		validate:       validate,
	}
}
