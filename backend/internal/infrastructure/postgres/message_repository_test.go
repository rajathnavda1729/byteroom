//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMessageRepo(t *testing.T) *MessageRepository {
	t.Helper()
	db := setupIntegrationDB(t)
	return NewMessageRepository(db)
}

func makeTestMessage(id, chatID, senderID string) *message.Message {
	return &message.Message{
		ID:          id,
		ChatID:      chatID,
		SenderID:    senderID,
		ContentType: message.ContentTypeMarkdown,
		Content:     "Hello world",
		CreatedAt:   time.Now().UTC(),
	}
}

func TestMessageRepository_Save_NewMessage_Persists(t *testing.T) {
	repo := setupMessageRepo(t)
	msg := makeTestMessage("msg-int-1", "chat-int-1", "user-int-1")

	err := repo.Save(context.Background(), msg)

	require.NoError(t, err)
}

func TestMessageRepository_Save_DuplicateID_IsIdempotent(t *testing.T) {
	repo := setupMessageRepo(t)
	msg := makeTestMessage("msg-idem-1", "chat-int-2", "user-int-1")
	require.NoError(t, repo.Save(context.Background(), msg))

	msg2 := makeTestMessage("msg-idem-1", "chat-int-2", "user-int-1")
	msg2.Content = "Different content"

	err := repo.Save(context.Background(), msg2)
	require.NoError(t, err)

	// Original content should be preserved
	found, err := repo.FindByID(context.Background(), "msg-idem-1")
	require.NoError(t, err)
	assert.Equal(t, "Hello world", found.Content)
}

func TestMessageRepository_FindByID_Exists_ReturnsMessage(t *testing.T) {
	repo := setupMessageRepo(t)
	msg := makeTestMessage("msg-find-1", "chat-int-3", "user-int-1")
	require.NoError(t, repo.Save(context.Background(), msg))

	found, err := repo.FindByID(context.Background(), msg.ID)

	require.NoError(t, err)
	assert.Equal(t, msg.Content, found.Content)
}

func TestMessageRepository_FindByID_NotExists_ReturnsError(t *testing.T) {
	repo := setupMessageRepo(t)

	_, err := repo.FindByID(context.Background(), "nonexistent-msg")

	assert.ErrorIs(t, err, message.ErrMessageNotFound)
}

func TestMessageRepository_FindByChatID_ReturnsPaginated(t *testing.T) {
	repo := setupMessageRepo(t)
	chatID := "chat-paginate-1"

	for i := 0; i < 15; i++ {
		msg := makeTestMessage(fmt.Sprintf("msg-page-%d", i), chatID, "user-int-1")
		require.NoError(t, repo.Save(context.Background(), msg))
	}

	msgs, err := repo.FindByChatID(context.Background(), chatID, 10, 0)

	require.NoError(t, err)
	assert.Len(t, msgs, 10)
}

func TestMessageRepository_ExistsByID_Exists_ReturnsTrue(t *testing.T) {
	repo := setupMessageRepo(t)
	msg := makeTestMessage("msg-exists-1", "chat-int-4", "user-int-1")
	require.NoError(t, repo.Save(context.Background(), msg))

	exists, err := repo.ExistsByID(context.Background(), msg.ID)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMessageRepository_ExistsByID_NotExists_ReturnsFalse(t *testing.T) {
	repo := setupMessageRepo(t)

	exists, err := repo.ExistsByID(context.Background(), "nonexistent-msg-x")

	require.NoError(t, err)
	assert.False(t, exists)
}
