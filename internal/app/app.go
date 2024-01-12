package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Run() {
	log := zap.Must(zap.NewProduction())
	defer log.Sync()

	log.Info("server started")

	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
