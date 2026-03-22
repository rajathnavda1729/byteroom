package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// MockRepository is a testify mock for the Repository interface.
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, u *User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) Search(ctx context.Context, query, excludeUserID string, limit int) ([]*User, error) {
	args := m.Called(ctx, query, excludeUserID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*User), args.Error(1)
}

func TestUserService_Register_ValidInput_CreatesUser(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	u, err := svc.Register(context.Background(), "alice", "Alice Smith", "s3cr3t!")

	require.NoError(t, err)
	assert.Equal(t, "alice", u.Username)
	assert.Equal(t, "Alice Smith", u.DisplayName)
	assert.NotEmpty(t, u.ID)
	assert.NotEmpty(t, u.PasswordHash)
	repo.AssertExpectations(t)
}

func TestUserService_Register_InvalidUsername_ReturnsError(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	_, err := svc.Register(context.Background(), "ab", "AB", "password")

	assert.ErrorIs(t, err, ErrInvalidUsername)
	repo.AssertNotCalled(t, "Save")
}

func TestUserService_Register_RepositoryError_PropagatesError(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*user.User")).
		Return(ErrDuplicateUsername)

	_, err := svc.Register(context.Background(), "alice", "Alice", "s3cr3t!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrDuplicateUsername)
}

func TestUserService_Authenticate_ValidCredentials_ReturnsUser(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("s3cr3t!"), bcrypt.MinCost)
	stored := &User{ID: "user-1", Username: "alice", PasswordHash: string(hash)}

	repo.On("FindByUsername", mock.Anything, "alice").Return(stored, nil)

	u, err := svc.Authenticate(context.Background(), "alice", "s3cr3t!")

	require.NoError(t, err)
	assert.Equal(t, "user-1", u.ID)
}

func TestUserService_Authenticate_WrongPassword_ReturnsError(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	stored := &User{ID: "user-1", Username: "alice", PasswordHash: string(hash)}

	repo.On("FindByUsername", mock.Anything, "alice").Return(stored, nil)

	_, err := svc.Authenticate(context.Background(), "alice", "wrong")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestUserService_Authenticate_UserNotFound_ReturnsInvalidCredentials(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("FindByUsername", mock.Anything, "ghost").
		Return(nil, errors.New("not found"))

	_, err := svc.Authenticate(context.Background(), "ghost", "pass")

	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestUserEntity_Validate_TooShortUsername_ReturnsError(t *testing.T) {
	u := &User{Username: "ab"}
	assert.ErrorIs(t, u.Validate(), ErrInvalidUsername)
}

func TestUserEntity_Validate_TooLongUsername_ReturnsError(t *testing.T) {
	u := &User{Username: "this_username_is_way_too_long_to_be_valid_xyz"}
	assert.ErrorIs(t, u.Validate(), ErrInvalidUsername)
}

func TestUserEntity_Validate_InvalidCharacters_ReturnsError(t *testing.T) {
	u := &User{Username: "alice@world"}
	assert.ErrorIs(t, u.Validate(), ErrInvalidUsername)
}

func TestUserEntity_Validate_ValidUsername_ReturnsNil(t *testing.T) {
	u := &User{Username: "alice_123"}
	assert.NoError(t, u.Validate())
}
