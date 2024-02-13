package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type OrderRepository interface {
	AddOrder(ctx context.Context, orderInfo *entity.OrderInfo) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type OrderService struct {
	orderRepository OrderRepository
}

func NewOrderService(orderRepository OrderRepository) *OrderService {
	return &OrderService{
		orderRepository: orderRepository,
	}
}

func (s *OrderService) AddOrder(ctx context.Context, orderInfo *entity.OrderInfo) (*entity.Order, error) {
	orderUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	orderInfo.ID = orderUUID.String()

	order, err := s.orderRepository.AddOrder(ctx, orderInfo)
	if errors.Is(err, entity.ErrOrderUniqueViolation) {
		if order.UserID == orderInfo.UserID {
			return nil, entity.ErrUploadedByThisUser
		}
		return nil, entity.ErrUploadedByAnotherUser
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrders(ctx context.Context, userID string) ([]entity.Order, error) {
	orders, err := s.orderRepository.GetOrders(ctx, userID)
	if err != nil {
		return nil, err
	}

	return orders, nil
}
