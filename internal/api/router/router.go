package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/api/middleware/decompress"
	"github.com/ivas1ly/gophermart/internal/api/middleware/reqlogger"
	"github.com/ivas1ly/gophermart/internal/config"
)

func NewRouter(cfg config.HTTP, log *zap.Logger) *chi.Mux {
	log.Info("init new router")
	router := chi.NewRouter()

	router.Use(
		reqlogger.New(log),
		middleware.Recoverer,
		middleware.Compress(cfg.CompressLevel),
		decompress.New(log),
	)

	return router
}
