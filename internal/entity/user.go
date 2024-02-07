package entity

import "time"

type User struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	ID        string
	Username  string
	Hash      string
	Balance   int64
}
