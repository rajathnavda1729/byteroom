//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserRepo(t *testing.T) *UserRepository {
	t.Helper()
	db := setupIntegrationDB(t)
	return NewUserRepository(db)
}

func makeTestUser(username string) *user.User {
	return &user.User{
		ID:           "test-user-" + username,
		Username:     username,
		DisplayName:  "Test " + username,
		AvatarURL:    "",
		PasswordHash: "$2a$10$dummy-hash",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
}

func TestUserRepository_Save_ValidUser_Persists(t *testing.T) {
	repo := setupUserRepo(t)
	u := makeTestUser("alice_int")

	err := repo.Save(context.Background(), u)

	require.NoError(t, err)
}

func TestUserRepository_Save_DuplicateUsername_ReturnsError(t *testing.T) {
	repo := setupUserRepo(t)
	u := makeTestUser("dupuser")
	require.NoError(t, repo.Save(context.Background(), u))

	u2 := makeTestUser("dupuser")
	u2.ID = "other-id"
	err := repo.Save(context.Background(), u2)

	assert.ErrorIs(t, err, user.ErrDuplicateUsername)
}

func TestUserRepository_FindByID_Exists_ReturnsUser(t *testing.T) {
	repo := setupUserRepo(t)
	u := makeTestUser("findbyid")
	require.NoError(t, repo.Save(context.Background(), u))

	found, err := repo.FindByID(context.Background(), u.ID)

	require.NoError(t, err)
	assert.Equal(t, u.ID, found.ID)
	assert.Equal(t, u.Username, found.Username)
}

func TestUserRepository_FindByID_NotExists_ReturnsError(t *testing.T) {
	repo := setupUserRepo(t)

	_, err := repo.FindByID(context.Background(), "nonexistent-id")

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}

func TestUserRepository_FindByUsername_Exists_ReturnsUser(t *testing.T) {
	repo := setupUserRepo(t)
	u := makeTestUser("findbyname")
	require.NoError(t, repo.Save(context.Background(), u))

	found, err := repo.FindByUsername(context.Background(), u.Username)

	require.NoError(t, err)
	assert.Equal(t, u.ID, found.ID)
}

func TestUserRepository_FindByUsername_NotExists_ReturnsError(t *testing.T) {
	repo := setupUserRepo(t)

	_, err := repo.FindByUsername(context.Background(), "ghost_user")

	assert.ErrorIs(t, err, user.ErrUserNotFound)
}
