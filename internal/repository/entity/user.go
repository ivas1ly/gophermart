package entity

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type User struct {
	ID        string
	Username  string
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt pgtype.Timestamptz
}

func ToUserFromRepo(user *User) *entity.User {
	var deletedAt *time.Time
	if user.DeletedAt.Valid {
		deletedAt = &user.DeletedAt.Time
	}

	return &entity.User{
		ID:        user.ID,
		Username:  user.Username,
		Hash:      user.Hash,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
