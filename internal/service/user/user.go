package user

import (
	"context"

	"github.com/ivas1ly/gophermart/internal/repository"
	def "github.com/ivas1ly/gophermart/internal/service"
)

var _ def.UserService = (*Service)(nil)

type Service struct {
	userRepository repository.UserRepository
}

func NewService(userRepository repository.UserRepository) *Service {
	return &Service{
		userRepository: userRepository,
	}
}

func (s *Service) Create(ctx context.Context) error {
	return nil
}
