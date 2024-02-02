package dto

import (
	"github.com/shopspring/decimal"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type BalanceResponse struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func ToUserBalanceResponse(userBalance *entity.UserBalance) *BalanceResponse {
	decimal.MarshalJSONWithoutQuotes = true

	divValue := decimal.NewFromInt(DecimalPartDiv)

	decimalBalance := decimal.NewFromInt(userBalance.Balance).Div(divValue)
	decimalWithdrawn := decimal.NewFromInt(userBalance.Withdrawn).Div(divValue)

	return &BalanceResponse{
		Balance:   decimalBalance,
		Withdrawn: decimalWithdrawn,
	}
}
