package user

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/internal/repository"
	def "github.com/ivas1ly/gophermart/internal/service"
	"github.com/ivas1ly/gophermart/internal/utils/argon2id"
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

func (s *Service) Register(ctx context.Context, username, password string) (*entity.User, error) {
	userUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.Create(ctx, userUUID.String(), username, hash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.userRepository.Find(ctx, username)
	if err != nil {
		return nil, err
	}

	ok, err := argon2id.ComparePasswordAndHash(password, user.Hash)
	if !ok {
		return nil, entity.ErrIncorrectLoginOrPassword
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}
