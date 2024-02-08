package jwt

import (
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToken(t *testing.T) {
	SigningKey = []byte("36626d331c8c44f2d72f348f36323743598e267e86b3e4aca27c5b433247ea72")

	id, err := uuid.NewV7()
	assert.NoError(t, err)

	t.Run("check token", func(t *testing.T) {
		signedToken, err := NewToken(SigningKey, id.String())
		assert.NoError(t, err)
		assert.Equal(t, len(strings.Split(signedToken, ".")), 3)

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(signedToken, claims, func(token *jwt.Token) (interface{}, error) {
			return SigningKey, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, token.Valid, true)
		assert.Equal(t, claims["sub"], id.String())
	})
}
