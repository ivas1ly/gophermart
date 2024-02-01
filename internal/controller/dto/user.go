package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserRequest struct {
	Username string `json:"login" validate:"required,gte=2,lte=255"`
	Password string `json:"password" validate:"required,gt=8,lte=1000"`
}

func ToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

type BalanceResponse struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func ToUserBalanceResponse(userBalance *entity.UserBalance) *BalanceResponse {
	decimal.MarshalJSONWithoutQuotes = true

	divModValue := decimal.NewFromInt(DecimalPartDivMod)

	decimalBalance := decimal.NewFromInt(userBalance.Balance).Div(divModValue)
	decimalWithdrawn := decimal.NewFromInt(userBalance.Withdrawn).Div(divModValue)

	return &BalanceResponse{
		Balance:   decimalBalance,
		Withdrawn: decimalWithdrawn,
	}
}
