package entity

import "errors"

var (
	ErrUsernameUniqueViolation  = errors.New("username already exists")
	ErrUsernameNotFound         = errors.New("username not found")
	ErrIncorrectLoginOrPassword = errors.New("incorrect login or password")
)
