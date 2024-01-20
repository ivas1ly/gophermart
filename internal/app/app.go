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

type App struct {
	log    *zap.Logger
	router *chi.Mux
	db     *postgres.DB
	cfg    config.Config
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	a := &App{}

	a.log = logger.New(cfg.App.LogLevel, logger.NewDefaultLoggerConfig()).
		With(zap.String("app", "gophermart"))

	a.cfg = cfg

	a.log.Info("init the database pool")
	db, err := postgres.New(ctx, cfg.DatabaseURI, cfg.DatabaseConnAttempts, cfg.DatabaseConnTimeout)
	if err != nil {
		a.log.Error("can't create pgx pool", zap.Error(err))
		return nil, err
	}

	a.log.Info("database connection established")
	a.db = db

	a.log.Info("init new router")
	a.router = chi.NewRouter()

	a.log.Info("init user service")
	usp := newUserServiceProvider(a.log)
	usp.UserRepository(db)
	usp.UserHandler().Register(a.router)

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	if a.db.Pool != nil {
		defer a.db.Pool.Close()
	}

	err := a.startHTTP(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) startHTTP(ctx context.Context) error {
	a.log.Info("start http server", zap.String("addr", a.cfg.RunAddress))

	err := http.ListenAndServe(a.cfg.RunAddress, a.router)
	if err != nil {
		panic(err)
	}

	return nil
}
