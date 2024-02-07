package entity

import "time"

type Balance struct {
	ID        string
	Balance   int64
	Withdrawn int64
}

type Withdraw struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	ID          string
	UserID      string
	OrderNumber string
	Withdrawn   int64
}
