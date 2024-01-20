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
	log            *zap.Logger
	userHandler    *user.Handler
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
		s.userRepository = userRepository.NewRepository(db, s.log)
	}
	return s.userRepository
}

func (s *userServiceProvider) UserService(ur repository.UserRepository) service.UserService {
	if s.userService == nil {
		s.userService = userService.NewService(ur, s.log)
	}
	return s.userService
}

func (s *userServiceProvider) UserHandler() *user.Handler {
	if s.userHandler == nil {
		s.userHandler = user.NewHandler(s.UserService(s.userRepository), s.log)
	}
	return s.userHandler
}
