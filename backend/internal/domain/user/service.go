package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service contains business logic for user management.
type Service struct {
	repo Repository
}

// NewService creates a new user Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, username, displayName, password string) (*User, error) {
	u := &User{
		ID:          uuid.New().String(),
		Username:    username,
		DisplayName: displayName,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}
	u.PasswordHash = string(hash)

	if err := s.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("saving user: %w", err)
	}

	return u, nil
}

// Authenticate verifies credentials and returns the user on success.
func (s *Service) Authenticate(ctx context.Context, username, password string) (*User, error) {
	u, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return u, nil
}

// GetByID retrieves a user by their ID.
func (s *Service) GetByID(ctx context.Context, id string) (*User, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return u, nil
}

// Search finds users whose username or display name contains the query,
// excluding the caller's own account. Returns at most limit results.
func (s *Service) Search(ctx context.Context, query, excludeID string, limit int) ([]*User, error) {
	if query == "" {
		return nil, nil
	}
	return s.repo.Search(ctx, query, excludeID, limit)
}
