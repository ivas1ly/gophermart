package reqlogger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func New(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "logger"))

		l.Info("added logger middleware")

		logFn := func(w http.ResponseWriter, r *http.Request) {
			entry := l.With(
				zap.String("uri", r.RequestURI),
				zap.String("method", r.Method),
			)

			buf, err := io.ReadAll(r.Body)
			if err != nil {
				l.Info("can't read body")
			} else {
				entry = entry.With(zap.String("body", string(buf)))
			}

			reader := io.NopCloser(bytes.NewBuffer(buf))
			r.Body = reader

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()

			defer func() {
				entry.Info("request",
					zap.String("duration", time.Since(start).String()),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(logFn)
	}
}
