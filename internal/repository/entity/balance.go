package entity

import "github.com/ivas1ly/gophermart/internal/entity"

type UserBalance struct {
	ID        string
	Balance   int64
	Withdrawn int64
}

func ToUserBalanceFromRepo(userBalance *UserBalance) *entity.UserBalance {
	return &entity.UserBalance{
		ID:        userBalance.ID,
		Balance:   userBalance.Balance,
		Withdrawn: userBalance.Withdrawn,
	}
}
