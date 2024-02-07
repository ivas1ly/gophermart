package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
)

const (
	attempts = 3
	divValue = 100
)

type Client interface {
	Get(url string) ([]byte, int, error)
	CanDoRequest() (*time.Time, error)
}

type Service interface {
	GetNewOrders(ctx context.Context) ([]entity.Order, error)
	UpdateOrders(ctx context.Context, orders ...entity.Order) error
}

type Worker struct {
	ws               Service
	client           Client
	log              *zap.Logger
	accrualSystemURL string
	pollInterval     time.Duration
}

func NewWorker(client Client, userService Service, accrualSystemURL string,
	pollInterval time.Duration, log *zap.Logger) *Worker {
	return &Worker{
		client:           client,
		ws:               userService,
		pollInterval:     pollInterval,
		accrualSystemURL: accrualSystemURL,
		log:              log.With(zap.String("worker", "accrual system")),
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.log.Info("start worker")

	inputCh, ticker := w.getNewOrders(ctx)
	defer ticker.Stop()

	w.updateOrder(ctx, w.getOrderAccrual(ctx, inputCh))

	<-ctx.Done()
}

func (w *Worker) getNewOrders(ctx context.Context) (chan []entity.Order, *time.Ticker) {
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
				orders, err := w.ws.GetNewOrders(ctx)
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

type response struct {
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func (w *Worker) getOrderAccrual(ctx context.Context, inputCh chan []entity.Order) chan []entity.Order {
	w.log.Info("start process order accrual")

	result := make(chan []entity.Order)

	go func() {
		defer close(result)

		for orders := range inputCh {
			processedOrders := make([]entity.Order, 0)

			for _, order := range orders {
				order := order

				addr := fmt.Sprintf("%s/api/orders/%s", w.accrualSystemURL, order.Number)

				w.log.Info("accrual system", zap.String("addr", addr))

				var resp []byte
				var status int
				var err error

				for i := 0; i < attempts; i++ {
					resp, status, err = w.client.Get(addr)
					if err != nil {
						w.log.Warn("can't make request", zap.Error(err))
						continue
					}

					var after *time.Time
					after, err = w.client.CanDoRequest()
					if err != nil {
						w.log.Info("retry check at", zap.String("time", err.Error()))
						time.Sleep(time.Until(*after))
						continue
					}

					continue
				}
				if err != nil {
					w.log.Info("response error, skip order", zap.Error(err))
					continue
				}

				var res response
				if status == http.StatusOK {
					err = json.Unmarshal(resp, &res)
					if err != nil {
						w.log.Info("can't unmarshal json, skip order", zap.Error(err))
						continue
					}
				}
				if status == http.StatusNoContent {
					w.log.Info("no content status, skip order")
					continue
				}

				w.log.Info("received order status from accrual system")
				if res.Status == entity.StatusProcessed.String() || res.Status == entity.StatusInvalid.String() {
					order.Accrual = res.Accrual.Mul(decimal.NewFromInt(divValue)).IntPart()
					order.Status = res.Status

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

func (w *Worker) updateOrder(ctx context.Context, inputCh chan []entity.Order) {
	w.log.Info("start update order status")

	go func() {
		for orders := range inputCh {
			select {
			case <-ctx.Done():
				w.log.Info("received done context")
				return
			default:
				if len(orders) == 0 {
					w.log.Info("nothing to update in database")
					continue
				}

				w.log.Info("trying to update orders status")
				err := w.ws.UpdateOrders(ctx, orders...)
				if err != nil {
					w.log.Info("can't update orders, skip")
					continue
				}
				w.log.Info("all orders updated")
			}
		}
	}()
}
