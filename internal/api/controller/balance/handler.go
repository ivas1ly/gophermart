package controller

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/jwt"
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
	tokenAuth      *jwtauth.JWTAuth
}

func NewBalanceHandler(balanceService BalanceService, validate *validator.Validate) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
		log:            zap.L().With(zap.String("handler", "balance")),
		validate:       validate,
		tokenAuth:      jwtauth.New("HS256", jwt.SigningKey, nil),
	}
}

func (bh *BalanceHandler) Register(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(bh.tokenAuth), jwtauth.Authenticator(bh.tokenAuth))
		r.Route("/balance", func(r chi.Router) {
			r.Get("/", bh.balance)
			r.Post("/withdraw", bh.withdraw)
		})
		r.Get("/withdrawals", bh.withdrawals)
	})
}
