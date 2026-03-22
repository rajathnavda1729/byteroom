package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT_Generate_ValidUser_ReturnsToken(t *testing.T) {
	mgr := NewJWTManager("secret-key", 24*time.Hour)

	token, err := mgr.Generate("user-123")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWT_Validate_ValidToken_ReturnsUserID(t *testing.T) {
	mgr := NewJWTManager("secret-key", 24*time.Hour)
	token, _ := mgr.Generate("user-123")

	userID, err := mgr.Validate(token)

	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)
}

func TestJWT_Validate_ExpiredToken_ReturnsError(t *testing.T) {
	mgr := NewJWTManager("secret-key", -1*time.Hour)
	token, _ := mgr.Generate("user-123")

	_, err := mgr.Validate(token)

	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestJWT_Validate_TamperedToken_ReturnsError(t *testing.T) {
	mgr := NewJWTManager("secret-key", 24*time.Hour)
	token, _ := mgr.Generate("user-123")

	_, err := mgr.Validate(token + "tampered")

	assert.ErrorIs(t, err, ErrTokenInvalid)
}

func TestJWT_Validate_WrongSecret_ReturnsError(t *testing.T) {
	mgr1 := NewJWTManager("secret-A", 24*time.Hour)
	mgr2 := NewJWTManager("secret-B", 24*time.Hour)

	token, _ := mgr1.Generate("user-123")
	_, err := mgr2.Validate(token)

	assert.ErrorIs(t, err, ErrTokenInvalid)
}

func TestJWT_Validate_EmptyToken_ReturnsError(t *testing.T) {
	mgr := NewJWTManager("secret-key", 24*time.Hour)

	_, err := mgr.Validate("")

	assert.ErrorIs(t, err, ErrTokenInvalid)
}
