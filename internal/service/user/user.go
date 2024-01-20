package user

import (
	"context"

	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/repository"
	def "github.com/ivas1ly/gophermart/internal/service"
)

var _ def.UserService = (*Service)(nil)

type Service struct {
	userRepository repository.UserRepository
	log            *zap.Logger
}

func NewService(userRepository repository.UserRepository, log *zap.Logger) *Service {
	return &Service{
		userRepository: userRepository,
		log:            log,
	}
}

func (s *Service) Create(ctx context.Context) error {
	return nil
}
