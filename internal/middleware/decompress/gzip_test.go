package decompress

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestGzipMiddleware(t *testing.T) {
	log := logger.New(defaultLogLevel, zap.NewDevelopmentConfig())

	r := chi.NewRouter()
	r.Use(New(log))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		_, _ = w.Write(b)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	testData := "Test Data"

	t.Run("can decompress body", func(t *testing.T) {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err := gz.Write([]byte(testData))
		assert.NoError(t, err)
		err = gz.Close()
		assert.NoError(t, err)

		resp, respBody := testRequest(t, ts, http.MethodGet, "/", "gzip", bytes.NewBuffer(buf.Bytes()))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, testData, respBody)
	})

	t.Run("can't decompress body", func(t *testing.T) {
		resp, respBody := testRequest(t, ts, http.MethodGet, "/", "gzip", bytes.NewBuffer([]byte(testData)))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Equal(t, `{"message":"can't decompress body"}`, strings.TrimSpace(respBody))
	})

	t.Run("without header", func(t *testing.T) {
		resp, respBody := testRequest(t, ts, http.MethodGet, "/", "", bytes.NewBuffer([]byte(testData)))
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, testData, respBody)
	})
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, header string,
	body io.Reader) (*http.Response, string) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTestClientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, ts.URL+path, body)
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", header)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
