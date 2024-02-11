package reqlogger

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/lib/logger"
)

const (
	defaultLogLevel          = "info"
	defaultTestClientTimeout = 3 * time.Second
)

func TestLoggerMiddlewaware(t *testing.T) {
	log := logger.New(defaultLogLevel, zap.NewDevelopmentConfig())

	r := chi.NewRouter()
	r.Use(New(log))

	testText := "test"
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testText))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, respBody := testRequest(t, ts, http.MethodGet, "/", nil)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testText, respBody)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
