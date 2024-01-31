package entity

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type Order struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt pgtype.Timestamptz
	ID        string
	UserID    string
	Number    string
	Status    string
	Accrual   int64
}

func ToOrderFromRepo(order *Order) *entity.Order {
	var deletedAt *time.Time
	if order.DeletedAt.Valid {
		deletedAt = &order.DeletedAt.Time
	}

	return &entity.Order{
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
		DeletedAt: deletedAt,
		Accrual:   order.Accrual,
		ID:        order.ID,
		UserID:    order.UserID,
		Number:    order.Number,
		Status:    order.Status,
	}
}

func ToOrdersFromRepo(orders []Order) []entity.Order {
	entities := make([]entity.Order, 0, len(orders))
	var deletedAt *time.Time

	for _, order := range orders {
		deleted := order.DeletedAt
		if order.DeletedAt.Valid {
			deletedAt = &deleted.Time
		}

		entities = append(entities, entity.Order{
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
			DeletedAt: deletedAt,
			ID:        order.ID,
			UserID:    order.UserID,
			Number:    order.Number,
			Status:    order.Status,
			Accrual:   order.Accrual,
		})
	}

	return entities
}
