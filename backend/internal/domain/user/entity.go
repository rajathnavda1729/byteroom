package user

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrDuplicateUsername  = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUsername    = errors.New("username must be 3-32 characters, alphanumeric and underscores only")
)

// User represents a registered user in the system.
type User struct {
	ID          string    `json:"user_id"      db:"id"`
	Username    string    `json:"username"     db:"username"`
	DisplayName string    `json:"display_name" db:"display_name"`
	AvatarURL   string    `json:"avatar_url"   db:"avatar_url"`
	PasswordHash string   `json:"-"            db:"password_hash"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"   db:"updated_at"`
}

// Validate checks that the user fields are well-formed.
func (u *User) Validate() error {
	if len(u.Username) < 3 || len(u.Username) > 32 {
		return ErrInvalidUsername
	}
	for _, c := range u.Username {
		if !isAlphanumericOrUnderscore(c) {
			return ErrInvalidUsername
		}
	}
	return nil
}

func isAlphanumericOrUnderscore(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}
