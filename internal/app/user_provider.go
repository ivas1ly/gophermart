package app

import (
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/controller/user"
	"github.com/ivas1ly/gophermart/internal/lib/storage/postgres"
	"github.com/ivas1ly/gophermart/internal/repository"
	userRepository "github.com/ivas1ly/gophermart/internal/repository/user"
	"github.com/ivas1ly/gophermart/internal/service"
	userService "github.com/ivas1ly/gophermart/internal/service/user"
)

type userServiceProvider struct {
	userRepository repository.UserRepository
	userService    service.UserService
	userHandler    *user.Handler
}

func newUserServiceProvider() *userServiceProvider {
	return &userServiceProvider{}
}

func (s *userServiceProvider) UserRepository(db *postgres.DB) repository.UserRepository {
	if s.userRepository == nil {
		s.userRepository = userRepository.NewRepository(db)
	}
	return s.userRepository
}

func (s *userServiceProvider) UserService(ur repository.UserRepository) service.UserService {
	if s.userService == nil {
		s.userService = userService.NewService(ur)
	}
	return s.userService
}

func (s *userServiceProvider) UserHandler(log *zap.Logger) *user.Handler {
	if s.userHandler == nil {
		s.userHandler = user.NewHandler(s.UserService(s.userRepository), log)
	}
	return s.userHandler
}
