package entity

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ivas1ly/gophermart/internal/entity"
)

type Balance struct {
	ID        string
	Balance   int64
	Withdrawn int64
}

func ToUserBalanceFromRepo(userBalance *Balance) *entity.Balance {
	return &entity.Balance{
		ID:        userBalance.ID,
		Balance:   userBalance.Balance,
		Withdrawn: userBalance.Withdrawn,
	}
}

type Withdraw struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   pgtype.Timestamptz
	ID          string
	UserID      string
	OrderNumber string
	Withdrawn   int64
}

func ToWithdrawalsFromRepo(withdrawals []Withdraw) []entity.Withdraw {
	entities := make([]entity.Withdraw, 0, len(withdrawals))
	var deletedAt *time.Time

	for _, withdraw := range withdrawals {
		deleted := withdraw.DeletedAt
		if withdraw.DeletedAt.Valid {
			deletedAt = &deleted.Time
		}

		entities = append(entities, entity.Withdraw{
			CreatedAt:   withdraw.CreatedAt,
			UpdatedAt:   withdraw.UpdatedAt,
			DeletedAt:   deletedAt,
			ID:          withdraw.ID,
			UserID:      withdraw.UserID,
			OrderNumber: withdraw.OrderNumber,
			Withdrawn:   withdraw.Withdrawn,
		})
	}

	return entities
}
