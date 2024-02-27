package controller

import (
	"context"

	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type OrderService interface {
	AddOrder(ctx context.Context, orderInfo *entity.OrderInfo) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type OrderHandler struct {
	orderService OrderService
	log          *zap.Logger
}

func NewOrderHandler(orderService OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		log:          zap.L().With(zap.String("handler", "order")),
	}
}
