package app

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/config"
	"github.com/ivas1ly/gophermart/internal/lib/logger"
	"github.com/ivas1ly/gophermart/internal/lib/migrate"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	"github.com/ivas1ly/gophermart/internal/middleware/reqlogger"
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
	db, err := postgres.New(ctx, cfg.DatabaseURI, cfg.DatabaseConnAttempts, cfg.DatabaseConnTimeout, a.log)
	if err != nil {
		a.log.Error("can't create pgx pool", zap.Error(err))
		return nil, err
	}

	a.log.Info("database connection established")
	a.db = db

	a.log.Info("trying to up migrations")
	err = migrate.Run(ctx, db.Pool)
	if err != nil {
		a.log.Info("can't run migrations", zap.Error(err))
		return nil, err
	}
	a.log.Info("migrations up success")

	a.log.Info("init new router")
	a.router = chi.NewRouter()
	a.router.Use(
		reqlogger.New(a.log),
		middleware.Recoverer,
		middleware.Compress(cfg.CompressLevel),
	)

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
