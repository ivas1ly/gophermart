package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	attempts = 3
	divValue = 100
)

type AccrualClient struct {
	log              *zap.Logger
	retryAt          *time.Time
	accrualSystemURL string
	clientTimeout    time.Duration
}

func NewAccrualClient(accrualSystemURL string, timeout time.Duration, log *zap.Logger) *AccrualClient {
	return &AccrualClient{
		log:              log.With(zap.String("client", "http")),
		clientTimeout:    timeout,
		accrualSystemURL: accrualSystemURL,
	}
}

type OrderStatusResponse struct {
	Status  string          `json:"status"`
	Accrual decimal.Decimal `json:"accrual"`
}

func (ac *AccrualClient) GetOrderStatus(id string) (string, int64, error) {
	addr := fmt.Sprintf("%s/api/orders/%s", ac.accrualSystemURL, id)

	var response *http.Response
	var err error

	for i := 0; i < attempts; i++ {
		response, err = ac.get(addr)
		if err != nil {
			ac.log.Warn("can't make request", zap.Error(err))
			response.Body.Close()
			continue
		}

		var after *time.Time
		after, err = ac.canDoRequest()
		if err != nil {
			ac.log.Info("retry check at", zap.String("time", err.Error()))
			time.Sleep(time.Until(*after))
			response.Body.Close()
			continue
		}
		continue
	}
	defer response.Body.Close()

	err = ac.checkResponse(response)
	if err != nil {
		ac.log.Info("check result", zap.String("result", err.Error()))
		return "", 0, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		ac.log.Info("can't read response body")
		return "", 0, err
	}

	var res *OrderStatusResponse
	if response.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &res)
		if err != nil {
			ac.log.Info("can't unmarshal json, skip order", zap.Error(err))
			return "", 0, err
		}
	}
	if response.StatusCode == http.StatusNoContent {
		ac.log.Info("no content status, skip order")
		return "", 0, fmt.Errorf("no content in response")
	}

	ac.log.Info("order status", zap.String("status", fmt.Sprintf("%+v", res)))

	accrual := res.Accrual.Mul(decimal.NewFromInt(divValue)).IntPart()

	return res.Status, accrual, nil
}

func (ac *AccrualClient) get(url string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ac.clientTimeout)
	defer cancel()

	ac.log.Info("new request", zap.String("url", url), zap.Duration("timeout", ac.clientTimeout))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	ac.log.Info("trying to do request")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ac.log.Info("http client", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

func (ac *AccrualClient) checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusTooManyRequests {
		ac.log.Info("request", zap.String("status", resp.Status))

		retryAfter, err := time.ParseDuration(fmt.Sprintf("%ss", resp.Header.Get("Retry-After")))
		if err != nil {
			return err
		}

		*ac.retryAt = time.Now().Add(retryAfter)
	}

	return nil
}

func (ac *AccrualClient) canDoRequest() (*time.Time, error) {
	if ac.retryAt == nil {
		ac.log.Info("ok, can do request")
		return nil, nil
	}

	if time.Now().After(*ac.retryAt) {
		ac.log.Info("now time after retry time, reset retry time")

		ac.retryAt = nil
		return nil, nil
	}

	return ac.retryAt, fmt.Errorf("can't send request in %s", time.Until(*ac.retryAt).String())
}
