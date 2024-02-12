package controller

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type BalanceResponse struct {
	Balance   decimal.Decimal `json:"current"`
	Withdrawn decimal.Decimal `json:"withdrawn"`
}

func ToUserBalanceResponse(userBalance *entity.Balance) *BalanceResponse {
	decimal.MarshalJSONWithoutQuotes = true

	divValue := decimal.NewFromInt(entity.DecimalPartDiv)

	decimalBalance := decimal.NewFromInt(userBalance.Balance).Div(divValue)
	decimalWithdrawn := decimal.NewFromInt(userBalance.Withdrawn).Div(divValue)

	return &BalanceResponse{
		Balance:   decimalBalance,
		Withdrawn: decimalWithdrawn,
	}
}

type WithdrawRequest struct {
	Order string          `json:"order" validate:"required,gte=4,lte=255"`
	Sum   decimal.Decimal `json:"sum" validate:"required"`
}

type WithdrawResponse struct {
	ProcessedAt time.Time       `json:"processed_at"`
	Order       string          `json:"order"`
	Sum         decimal.Decimal `json:"sum"`
}

func ToWithdrawalsResponse(withdrawals []entity.Withdraw) []WithdrawResponse {
	entities := make([]WithdrawResponse, 0, len(withdrawals))

	decimal.MarshalJSONWithoutQuotes = true

	divValue := decimal.NewFromInt(entity.DecimalPartDiv)

	for _, withdraw := range withdrawals {
		response := WithdrawResponse{
			ProcessedAt: withdraw.CreatedAt,
			Order:       withdraw.OrderNumber,
			Sum:         decimal.NewFromInt(withdraw.Withdrawn).Div(divValue),
		}

		entities = append(entities, response)
	}

	return entities
}
