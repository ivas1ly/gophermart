package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/service"
)

type Handler struct {
	userService service.UserService
	log         *zap.Logger
}

func NewHandler(userService service.UserService, log *zap.Logger) *Handler {
	return &Handler{
		userService: userService,
		log:         log,
	}
}

func (h *Handler) Register(router *chi.Mux) {
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/login", h.login)
		r.Post("/register", h.register)
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("register"))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("login"))
}
