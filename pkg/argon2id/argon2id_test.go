package argon2id

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHash(t *testing.T) {
	hashRegExp := regexp.
		MustCompile(`^\$argon2id\$v=19\$m=65536,t=1,p=[0-9]{1,4}\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`)

	hash1, err := CreateHash("uwuowo", DefaultParams)
	assert.NoError(t, err)
	assert.True(t, hashRegExp.MatchString(hash1))

	hash2, err := CreateHash("uwuowo", DefaultParams)
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash2)
}

func TestComparePasswordAndHash(t *testing.T) {
	hash, err := CreateHash("uwuowo", DefaultParams)
	assert.NoError(t, err)

	match, err := ComparePasswordAndHash("uwuowo", hash)
	assert.NoError(t, err)
	assert.True(t, match)

	match, err = ComparePasswordAndHash("owouwu", hash)
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestDecodeHash(t *testing.T) {
	hash, err := CreateHash("uwuowo", DefaultParams)
	assert.NoError(t, err)

	params, _, _, err := DecodeHash(hash)
	assert.NoError(t, err)
	assert.Equal(t, *params, *DefaultParams)
}

func TestCheckHash(t *testing.T) {
	hash, err := CreateHash("uwuowo", DefaultParams)
	assert.NoError(t, err)

	ok, params, err := CheckHash("uwuowo", hash)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, *params, *DefaultParams)
}

func TestStrictDecoding(t *testing.T) {
	ok, _, err := CheckHash("qwerty123",
		"$argon2id$v=19$m=65536,t=1,p=8$QrnU5AAUZNdblXhW48VaQA$+FMwKJQ/VizkNgWDXFf9yPMLsomaXQDDT0OFbZEw9SU")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, _, err = CheckHash("qwerty123",
		"$argon2id$v=19$m=65536,t=1,p=8$QrnU5AAUZNdblXhW48VaQA$+FMwKJQ/VizkNgWDXFf9yPMLsomaXQDDT0OFbZEw9SF")
	assert.Error(t, err)
	assert.False(t, ok)
}

func TestVariant(t *testing.T) {
	_, _, err := CheckHash("qwerty123",
		"$argon2d$v=19$m=65536,t=1,p=8$QrnU5AAUZNdblXhW48VaQA$+FMwKJQ/VizkNgWDXFf9yPMLsomaXQDDT0OFbZEw9SU")
	assert.ErrorIs(t, err, ErrIncompatibleVariant)
}
