package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/argon2id"
)

type AuthRepository interface {
	Create(ctx context.Context, id, username, hash string) (*entity.User, error)
	Find(ctx context.Context, username string) (*entity.User, error)
}

type AuthService struct {
	authRepository AuthRepository
}

func NewAuthService(userRepository AuthRepository) *AuthService {
	return &AuthService{
		authRepository: userRepository,
	}
}

func (s *AuthService) Register(ctx context.Context, username, password string) (*entity.User, error) {
	userUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	user, err := s.authRepository.Create(ctx, userUUID.String(), username, hash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.authRepository.Find(ctx, username)
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
