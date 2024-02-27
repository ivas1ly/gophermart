package decompress

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"go.uber.org/zap"
)

func New(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		l := log.With(zap.String("middleware", "decompress"))

		l.Info("added decompress middleware")

		gzipFn := func(w http.ResponseWriter, r *http.Request) {
			ok := checkHasGzipEncoding(r.Header.Values("Content-Encoding"))
			if ok {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					l.Info("can't decompress body")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)

					render.JSON(w, r, render.M{"message": "can't decompress body"})
					return
				}

				r.Body = cr
				l.Info("body decompressed")
				defer cr.Close()
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(gzipFn)
	}
}

func checkHasGzipEncoding(values []string) bool {
	for _, value := range values {
		if value == "gzip" {
			return true
		}
	}
	return false
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
