package dto

import (
	"time"

	"github.com/shopspring/decimal"

	"github.com/ivas1ly/gophermart/internal/entity"
)

const (
	DecimalPartDivMod = 100
)

type OrderResponse struct {
	Number    string `json:"number"`
	Status    string `json:"status"`
	CreatedAt string `json:"uploaded_at"`
}

func ToOrderResponse(order *entity.Order) *OrderResponse {
	return &OrderResponse{
		Number:    order.Number,
		Status:    order.Status,
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}
}

type OrdersResponse struct {
	Accrual *decimal.Decimal `json:"accrual,omitempty"`
	OrderResponse
}

func ToOrdersResponse(orders []entity.Order) []OrdersResponse {
	entities := make([]OrdersResponse, 0, len(orders))

	decimal.MarshalJSONWithoutQuotes = true

	divModValue := decimal.NewFromInt(DecimalPartDivMod)

	for _, order := range orders {
		accrual := order.Accrual

		response := OrdersResponse{
			OrderResponse: OrderResponse{
				Number:    order.Number,
				Status:    order.Status,
				CreatedAt: order.CreatedAt.Format(time.RFC3339),
			},
		}

		if order.Status == entity.StatusProcessed.String() {
			decimalAccrual := decimal.NewFromInt(accrual).Div(divModValue)
			response.Accrual = &decimalAccrual
		}

		entities = append(entities, response)
	}

	return entities
}
