package controller

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/jwt"
)

type OrderService interface {
	AddOrder(ctx context.Context, userID, orderNumber string) (*entity.Order, error)
	GetOrders(ctx context.Context, userID string) ([]entity.Order, error)
}

type OrderHandler struct {
	orderService OrderService
	log          *zap.Logger
	tokenAuth    *jwtauth.JWTAuth
}

func NewOrderHandler(orderService OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		log:          zap.L().With(zap.String("handler", "order")),
		tokenAuth:    jwtauth.New("HS256", jwt.SigningKey, nil),
	}
}

func (oh *OrderHandler) Register(router chi.Router) {
	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(oh.tokenAuth), jwtauth.Authenticator(oh.tokenAuth))
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", oh.order)
			r.Get("/", oh.orders)
		})
	})
}
