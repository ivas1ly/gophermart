package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	auth "github.com/ivas1ly/gophermart/internal/api/controller/auth"
	balance "github.com/ivas1ly/gophermart/internal/api/controller/balance"
	order "github.com/ivas1ly/gophermart/internal/api/controller/order"
	"github.com/ivas1ly/gophermart/internal/app/provider"
	"github.com/ivas1ly/gophermart/pkg/jwt"
)

func RegisterRoutes(router *chi.Mux, sp *provider.ServiceProvider, validate *validator.Validate) {
	authHandler := auth.NewAuthHandler(sp.AuthService, validate)
	orderHandler := order.NewOrderHandler(sp.OrderService)
	balanceHandler := balance.NewBalanceHandler(sp.BalanceService, validate)

	tokenAuth := jwtauth.New("HS256", jwt.SigningKey, nil)

	zap.L().Info("register routes")
	router.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/register", authHandler.Register)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth), jwtauth.Authenticator(tokenAuth))
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.Order)
				r.Get("/", orderHandler.Orders)
			})

			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balanceHandler.Balance)
				r.Post("/withdraw", balanceHandler.Withdraw)
			})
			r.Get("/withdrawals", balanceHandler.Withdrawals)
		})
	})
}
