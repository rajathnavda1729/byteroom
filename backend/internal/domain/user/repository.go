package user

import "context"

// Repository defines persistence operations for users.
// Implemented by infrastructure/postgres.
type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Search(ctx context.Context, query string, excludeID string, limit int) ([]*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id string) error
}
