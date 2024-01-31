package entity

import "time"

type Order struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	ID        string
	UserID    string
	Number    string
	Status    string
	Accrual   int64
}

type Status int

const (
	StatusNew Status = iota
	StatusProcessing
	StatusInvalid
	StatusProcessed
)

func (s Status) String() string {
	switch s {
	case StatusNew:
		return "NEW"
	case StatusProcessing:
		return "PROCESSING"
	case StatusInvalid:
		return "INVALID"
	case StatusProcessed:
		return "PROCESSED"
	}
	return "UNKNOWN"
}
