package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	OneDay = 24 * time.Hour
)

var SigningKey []byte

func NewToken(key []byte, id string) (string, error) {
	tokenID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("jwt can't get new uuid v7: %w", err)
	}

	claims := jwt.RegisteredClaims{
		Issuer:    "gophermart",
		Subject:   id,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(OneDay)),
		NotBefore: jwt.NewNumericDate(time.Now()),
		ID:        tokenID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("jwt can't sign string: %w", err)
	}

	return ss, nil
}
