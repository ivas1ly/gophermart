package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/ivas1ly/gophermart/internal/entity"
	"github.com/ivas1ly/gophermart/pkg/argon2id"
)

type AuthRepository interface {
	AddUser(ctx context.Context, userInfo *entity.UserInfo) (*entity.User, error)
	FindUser(ctx context.Context, username string) (*entity.User, error)
}

type AuthService struct {
	authRepository AuthRepository
}

func NewAuthService(authRepository AuthRepository) *AuthService {
	return &AuthService{
		authRepository: authRepository,
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

	userInfo := &entity.UserInfo{
		ID:       userUUID.String(),
		Username: username,
		Hash:     hash,
	}

	user, err := s.authRepository.AddUser(ctx, userInfo)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*entity.User, error) {
	user, err := s.authRepository.FindUser(ctx, username)
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
