package entity

import "errors"

var (
	ErrUsernameUniqueViolation  = errors.New("username already exists")
	ErrUsernameNotFound         = errors.New("username not found")
	ErrIncorrectLoginOrPassword = errors.New("incorrect login or password")

	ErrOrderUniqueViolation  = errors.New("order already exists")
	ErrUploadedByThisUser    = errors.New("already uploaded by this user")
	ErrUploadedByAnotherUser = errors.New("already uploaded by another user")
	ErrNoOrdersFound         = errors.New("no orders found")

	ErrNotEnoughPointsToWithraw = errors.New("not enough points to withdraw")
)
