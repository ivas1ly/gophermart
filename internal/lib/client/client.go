package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type HTTP struct {
	retryAt       *time.Time
	log           *zap.Logger
	clientTimeout time.Duration
}

func NewClient(timeout time.Duration, log *zap.Logger) *HTTP {
	return &HTTP{
		log:           log.With(zap.String("client", "http")),
		clientTimeout: timeout,
	}
}

func (h *HTTP) Get(url string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), h.clientTimeout)
	defer cancel()

	h.log.Info("new request", zap.String("url", url), zap.Duration("timeout", h.clientTimeout))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	h.log.Info("trying to do request")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.log.Info("httt client", zap.Error(err))
		return nil, 0, err
	}
	defer resp.Body.Close()

	err = h.checkResponse(resp)
	if err != nil {
		h.log.Info("check result", zap.String("result", err.Error()))
		return nil, resp.StatusCode, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.log.Info("can't read response body")
		return nil, resp.StatusCode, err
	}

	return body, resp.StatusCode, nil
}

func (h *HTTP) checkResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusTooManyRequests {
		h.log.Info("request", zap.String("status", resp.Status))

		retryAfter, err := time.ParseDuration(fmt.Sprintf("%ss", resp.Header.Get("Retry-After")))
		if err != nil {
			return err
		}

		*h.retryAt = time.Now().Add(retryAfter)
	}

	return nil
}

func (h *HTTP) CanDoRequest() (*time.Time, error) {
	if h.retryAt == nil {
		h.log.Info("ok, can do request")
		return nil, nil
	}

	if time.Now().After(*h.retryAt) {
		h.log.Info("now time after retry time, reset retry time")

		h.retryAt = nil
		return nil, nil
	}

	return h.retryAt, fmt.Errorf("can't send request in %s", time.Until(*h.retryAt).String())
}
