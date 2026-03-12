package util

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	pass := RandomString(6)
	hash1, err := HashPassword(pass)
	require.NoError(t, err)
	require.NotEmpty(t, hash1)

	err = CheckPassword(pass, hash1)
	require.NoError(t, err)

	err = CheckPassword("wrong", hash1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hash2, err := HashPassword(pass)
	require.NoError(t, err)
	require.NotEmpty(t, hash2)
	require.NotEqual(t, hash1, hash2)
}
