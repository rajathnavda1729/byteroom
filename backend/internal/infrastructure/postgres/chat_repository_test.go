//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupChatRepo(t *testing.T) *ChatRepository {
	t.Helper()
	db := setupIntegrationDB(t)
	return NewChatRepository(db)
}

func makeTestChat(createdBy string) *chat.Chat {
	return &chat.Chat{
		ID:        "chat-int-" + createdBy,
		Name:      "Test Chat",
		Type:      chat.ChatTypeGroup,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Members:   []string{},
	}
}

func TestChatRepository_Save_GroupChat_Persists(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u1")

	err := repo.Save(context.Background(), c)

	require.NoError(t, err)
}

func TestChatRepository_FindByID_Exists_ReturnsChat(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u2")
	require.NoError(t, repo.Save(context.Background(), c))

	found, err := repo.FindByID(context.Background(), c.ID)

	require.NoError(t, err)
	assert.Equal(t, c.ID, found.ID)
	assert.Equal(t, c.Name, found.Name)
}

func TestChatRepository_FindByID_NotExists_ReturnsError(t *testing.T) {
	repo := setupChatRepo(t)

	_, err := repo.FindByID(context.Background(), "nonexistent-chat")

	assert.ErrorIs(t, err, chat.ErrChatNotFound)
}

func TestChatRepository_AddMember_Success(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u3")
	require.NoError(t, repo.Save(context.Background(), c))

	err := repo.AddMember(context.Background(), c.ID, "u3", chat.RoleAdmin)

	assert.NoError(t, err)
}

func TestChatRepository_AddMember_Idempotent(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u4")
	require.NoError(t, repo.Save(context.Background(), c))
	require.NoError(t, repo.AddMember(context.Background(), c.ID, "u4", chat.RoleMember))

	// Second add with same user should not error (ON CONFLICT DO NOTHING)
	err := repo.AddMember(context.Background(), c.ID, "u4", chat.RoleMember)

	assert.NoError(t, err)
}

func TestChatRepository_IsMember_AfterAdd_ReturnsTrue(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u5")
	require.NoError(t, repo.Save(context.Background(), c))
	require.NoError(t, repo.AddMember(context.Background(), c.ID, "u5", chat.RoleMember))

	ok, err := repo.IsMember(context.Background(), c.ID, "u5")

	require.NoError(t, err)
	assert.True(t, ok)
}

func TestChatRepository_IsMember_NotAdded_ReturnsFalse(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u6")
	require.NoError(t, repo.Save(context.Background(), c))

	ok, err := repo.IsMember(context.Background(), c.ID, "stranger")

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestChatRepository_GetMemberRole_Admin_ReturnsAdmin(t *testing.T) {
	repo := setupChatRepo(t)
	c := makeTestChat("u7")
	require.NoError(t, repo.Save(context.Background(), c))
	require.NoError(t, repo.AddMember(context.Background(), c.ID, "u7", chat.RoleAdmin))

	role, err := repo.GetMemberRole(context.Background(), c.ID, "u7")

	require.NoError(t, err)
	assert.Equal(t, chat.RoleAdmin, role)
}
