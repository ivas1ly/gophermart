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

	t.Run("correct number 2", func(t *testing.T) {
		ok := CheckNumber("41632078327500")
		assert.Equal(t, ok, true)
	})

	t.Run("correct number 3", func(t *testing.T) {
		ok := CheckNumber("45444541846")
		assert.Equal(t, ok, true)
	})

	t.Run("correct number 4", func(t *testing.T) {
		ok := CheckNumber("333207722682")
		assert.Equal(t, ok, true)
	})

	t.Run("correct number 5", func(t *testing.T) {
		ok := CheckNumber("836361")
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
