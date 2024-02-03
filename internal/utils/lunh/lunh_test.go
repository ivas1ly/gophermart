package lunh

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLunh(t *testing.T) {
	t.Run("correct number", func(t *testing.T) {
		ok := CheckNumber("2377225624")
		assert.Equal(t, ok, true)
	})

	t.Run("incorrect number", func(t *testing.T) {
		ok := CheckNumber("2377225621")
		assert.Equal(t, ok, false)
	})

	t.Run("simple string", func(t *testing.T) {
		ok := CheckNumber("test")
		assert.Equal(t, ok, false)
	})

	t.Run("string with space", func(t *testing.T) {
		ok := CheckNumber(" ")
		assert.Equal(t, ok, false)
	})

	t.Run("empty string", func(t *testing.T) {
		ok := CheckNumber("")
		assert.Equal(t, ok, false)
	})
}
