package worker

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type AccrualClient interface {
	GetOrderStatus(id string) (string, int64, error)
}

type AccrualService interface {
	GetNewOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrders(ctx context.Context, orders ...entity.Order) error
}

type AccrualWorker struct {
	as           AccrualService
	client       AccrualClient
	log          *zap.Logger
	pollInterval time.Duration
}

func NewAccrualWorker(accrualClient AccrualClient, accrualService AccrualService,
	pollInterval time.Duration, log *zap.Logger) *AccrualWorker {
	return &AccrualWorker{
		client:       accrualClient,
		as:           accrualService,
		pollInterval: pollInterval,
		log:          log.With(zap.String("worker", "accrual system")),
	}
}

func (w *AccrualWorker) Run(ctx context.Context) {
	w.log.Info("start worker")

	inputCh, ticker := w.getNewOrders(ctx)
	defer ticker.Stop()

	w.updateOrder(ctx, w.getOrderAccrual(ctx, inputCh))

	<-ctx.Done()
}

func (w *AccrualWorker) getNewOrders(ctx context.Context) (chan []entity.Order, *time.Ticker) {
	w.log.Info("start process orders with interval", zap.Duration("poll interval", w.pollInterval))

	updateTicker := time.NewTicker(w.pollInterval)

	inputCh := make(chan []entity.Order)

	go func() {
		defer close(inputCh)

		w.log.Debug("start polling new orders")

		for {
			select {
			case <-ctx.Done():
				w.log.Info("received done context")
				return
			case <-updateTicker.C:
				w.log.Info("trying to get new orders")
				orders, err := w.as.GetNewOrders(ctx)
				if err != nil {
					w.log.Info("can't get new orders", zap.Error(err))
					continue
				}
				if len(orders) == 0 {
					w.log.Info("no new orders to process")
					continue
				}

				w.log.Info("add orders to process queue")
				inputCh <- orders
			}
		}
	}()

	return inputCh, updateTicker
}

func (w *AccrualWorker) getOrderAccrual(ctx context.Context, inputCh chan []entity.Order) chan []entity.Order {
	w.log.Info("start process order accrual")

	result := make(chan []entity.Order)

	go func() {
		defer close(result)

		for orders := range inputCh {
			processedOrders := make([]entity.Order, 0)

			for _, order := range orders {
				order := order

				w.log.Info("check order", zap.String("order", order.Number))

				status, accrual, err := w.client.GetOrderStatus(order.Number)
				if err != nil {
					w.log.Info("response error, skip order", zap.Error(err))
					continue
				}

				w.log.Info("received order status from accrual system", zap.String("status", status))
				if status == entity.StatusProcessed.String() || status == entity.StatusInvalid.String() {
					order.Status = status
					order.Accrual = accrual

					w.log.Info("adding new order to queue", zap.String("order",
						fmt.Sprintf("%+v", order)))

					processedOrders = append(processedOrders, order)
					w.log.Info("order added to queue for status update")
				}
			}

			select {
			case <-ctx.Done():
				w.log.Info("received done context")
				return
			case result <- processedOrders:
				w.log.Info("orders have been pushed to the next stage for update")
			}
		}
	}()

	return result
}

func (w *AccrualWorker) updateOrder(ctx context.Context, inputCh chan []entity.Order) {
	w.log.Info("start update order status")

	go func() {
		for {
			select {
			case <-ctx.Done():
				w.log.Info("received done context")
				return
			case orders := <-inputCh:
				if len(orders) == 0 {
					w.log.Info("nothing to update in database")
					continue
				}

				w.log.Info("trying to update orders status")
				err := w.as.UpdateOrders(ctx, orders...)
				if err != nil {
					w.log.Info("can't update orders, skip")
					continue
				}
				w.log.Info("all orders updated")
			}
		}
	}()
}
