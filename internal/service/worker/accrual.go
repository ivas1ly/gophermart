package workers

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type AccrualWorkerRepository interface {
	GetOrdersToProcess(ctx context.Context) ([]entity.Order, error)
	UpdateOrderAndUserBalance(ctx context.Context, order entity.Order) error
}

type AccrualWorkerService struct {
	workerRepository AccrualWorkerRepository
}

func NewAccrualWorkerService(workerRepository AccrualWorkerRepository) *AccrualWorkerService {
	return &AccrualWorkerService{
		workerRepository: workerRepository,
	}
}

func (s *AccrualWorkerService) GetNewOrders(ctx context.Context) ([]entity.Order, error) {
	orders, err := s.workerRepository.GetOrdersToProcess(ctx)
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *AccrualWorkerService) UpdateOrders(ctx context.Context, orders ...entity.Order) error {
	zap.L().Info("updating order status and user balance")

	for _, order := range orders {
		err := s.workerRepository.UpdateOrderAndUserBalance(ctx, order)
		if errors.Is(err, entity.ErrCanNotUpdateOrder) {
			zap.L().Warn("can't update order status", zap.Error(err))
			continue
		}
		if errors.Is(err, entity.ErrCanNotUpdateUserBalance) {
			zap.L().Warn("can't update user balance", zap.Error(err))
			continue
		}
		if err != nil {
			zap.L().Warn("can't update order and user balance", zap.Error(err))
			continue
		}
	}

	zap.L().Info("order status and user balance updated")

	return nil
}
