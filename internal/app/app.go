package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/config"
	"github.com/ivas1ly/gophermart/internal/lib/logger"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
)

func Run(cfg config.Config) error {
	log := logger.New(cfg.App.LogLevel, logger.NewDefaultLoggerConfig()).
		With(zap.String("app", "gophermart"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, cfg.DatabaseURI, cfg.DatabaseConnAttempts, cfg.DatabaseConnTimeout)
	if err != nil {
		log.Error("can't create pgx pool", zap.Error(err))
		return err
	}
	defer db.Pool.Close()
	log.Info("database connection established")

	r := chi.NewRouter()

	usp := newUserServiceProvider()
	usp.UserRepository(db)
	usp.UserHandler(log).Register(r)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	err = http.ListenAndServe(cfg.RunAddress, r)
	if err != nil {
		panic(err)
	}

	return nil
}
