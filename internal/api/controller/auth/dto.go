package controller

import (
	"time"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserRequest struct {
	Username string `json:"login" validate:"required,gte=2,lte=255"`
	Password string `json:"password" validate:"required,gt=8,lte=1000"`
}

func ToUserResponse(user *entity.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
