package app

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/controller"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	"github.com/ivas1ly/gophermart/internal/repository"
	"github.com/ivas1ly/gophermart/internal/service"
)

type userServiceProvider struct {
	log            *zap.Logger
	userHandler    *controller.Handler
	userRepository repository.UserRepository
	userService    service.UserService
}

func newUserServiceProvider(log *zap.Logger) *userServiceProvider {
	return &userServiceProvider{
		log: log.With(zap.String("service", "user")),
	}
}

func (s *userServiceProvider) UserRepository(db *postgres.DB) repository.UserRepository {
	if s.userRepository == nil {
		s.userRepository = repository.NewRepository(db, s.log)
	}
	return s.userRepository
}

func (s *userServiceProvider) UserService(ur repository.UserRepository) service.UserService {
	if s.userService == nil {
		s.userService = service.NewService(ur, s.log)
	}
	return s.userService
}

func (s *userServiceProvider) UserHandler(validate *validator.Validate) *controller.Handler {
	if s.userHandler == nil {
		s.userHandler = controller.NewHandler(s.UserService(s.userRepository), validate, s.log)
	}
	return s.userHandler
}
