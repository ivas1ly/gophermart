package controller

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
)

const (
	AuthorizationSchema = "Bearer"
	AuthorizationHeader = "Authorization"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*entity.User, error)
	Login(ctx context.Context, username, password string) (*entity.User, error)
}

type AuthHandler struct {
	authService AuthService
	log         *zap.Logger
	validate    *validator.Validate
}

func NewAuthHandler(authService AuthService, validate *validator.Validate) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         zap.L().With(zap.String("handler", "auth")),
		validate:    validate,
	}
}
