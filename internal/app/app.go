package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/api/router"
	"github.com/ivas1ly/gophermart/internal/app/provider"
	"github.com/ivas1ly/gophermart/internal/config"
	"github.com/ivas1ly/gophermart/internal/lib/client"
	"github.com/ivas1ly/gophermart/internal/lib/logger"
	"github.com/ivas1ly/gophermart/internal/lib/migrate"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	"github.com/ivas1ly/gophermart/internal/repository"
	"github.com/ivas1ly/gophermart/internal/worker"
	"github.com/ivas1ly/gophermart/pkg/jwt"
)

type App struct {
	log    *zap.Logger
	router *chi.Mux
	db     *postgres.DB
	worker *worker.AccrualWorker
	cfg    config.Config
}

func NewApp(ctx context.Context, cfg config.Config) (*App, error) {
	log := logger.New(cfg.App.LogLevel, logger.NewDefaultLoggerConfig()).
		With(zap.String("app", "gophermart"))
	logger.SetGlobalLogger(log)

	a := &App{
		cfg: cfg,
		log: log,
	}
	jwt.SigningKey = cfg.SigningKey

	a.log.Info("init the database pool")
	db, err := postgres.New(ctx, cfg.DatabaseURI, cfg.DatabaseConnAttempts, cfg.DatabaseConnTimeout)
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

	a.router = router.NewRouter(cfg.HTTP, a.log)

	a.log.Info("init services")
	// "provider" name to avoid import cycle
	serviceProvider := provider.NewServiceProvider(db)
	serviceProvider.RegisterServices()

	a.log.Info("init api routes")
	validate := validator.New(validator.WithRequiredStructEnabled())
	router.RegisterRoutes(a.router, serviceProvider, validate)

	accrualClient := client.NewAccrualClient(cfg.AccrualSystemAddress, cfg.ClientTimeout, a.log)
	a.worker = worker.NewAccrualWorker(accrualClient, repository.NewAccrualWorkerRepository(a.db),
		cfg.WorkerPollInterval, a.log)

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	if a.db.Pool != nil {
		defer func() {
			a.log.Info("close database pool")
			a.db.Pool.Close()
		}()
	}

	notifyCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go a.worker.Run(ctx)

	if err := a.startHTTP(notifyCtx); err != nil {
		a.log.Error("unexpected server error", zap.Error(err))
	}

	a.log.Info("server was successfully shut down")

	return nil
}

func (a *App) startHTTP(ctx context.Context) error {
	server := &http.Server{
		Addr:              a.cfg.RunAddress,
		Handler:           a.router,
		ReadTimeout:       a.cfg.ReadTimeout,
		ReadHeaderTimeout: a.cfg.ReadHeaderTimeout,
		WriteTimeout:      a.cfg.WriteTimeout,
		IdleTimeout:       a.cfg.IdleTimeout,
	}

	err := chi.Walk(a.router, func(method string, route string, _ http.Handler,
		middlewares ...func(http.Handler) http.Handler) error {
		a.log.Info(fmt.Sprintf("[%s]: '%s' has %d middlewares", method, route, len(middlewares)))
		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.log.Error("unexpected server error", zap.Error(err))
		}
	}()

	a.log.Info("server started", zap.String("addr", a.cfg.RunAddress))
	<-ctx.Done()

	a.log.Info("gracefully shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()

	go func() {
		if err := server.Shutdown(shutdownCtx); err != nil {
			a.log.Error("unexpected server shutdown error", zap.Error(err))
		}
	}()

	<-shutdownCtx.Done()
	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		a.log.Info("timeout exceeded, forcing shutdown")
		return shutdownCtx.Err()
	}

	return nil
}
