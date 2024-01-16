package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/config"
)

func Run(cfg config.Config) {
	log := zap.Must(zap.NewProduction())
	defer log.Sync()

	log.Info("server started")

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	err := http.ListenAndServe(cfg.RunAddress, r)
	if err != nil {
		panic(err)
	}
}
