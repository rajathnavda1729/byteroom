//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/message"
	"github.com/byteroom/backend/internal/infrastructure/postgres"
	"github.com/google/uuid"
)

func BenchmarkMessageRepository_Save(b *testing.B) {
	db := setupIntegrationDB(b)
	repo := postgres.NewMessageRepository(db)
	ctx := context.Background()

	// Seed a chat
	_, err := db.ExecContext(ctx, `INSERT INTO chats (chat_id, chat_type) VALUES ($1, 'group') ON CONFLICT DO NOTHING`, "bench-chat")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		msg := &message.Message{
			ID:          uuid.New().String(),
			ChatID:      "bench-chat",
			SenderID:    "bench-user",
			ContentType: "markdown",
			Content:     "Benchmark message content",
			Timestamp:   time.Now(),
		}
		if err := repo.Save(ctx, msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMessageRepository_FindByChatID(b *testing.B) {
	db := setupIntegrationDB(b)
	repo := postgres.NewMessageRepository(db)
	ctx := context.Background()

	// Seed chat and messages
	chatID := fmt.Sprintf("bench-find-%s", uuid.New().String()[:8])
	_, _ = db.ExecContext(ctx, `INSERT INTO chats (chat_id, chat_type) VALUES ($1, 'group')`, chatID)

	for i := range 1000 {
		_ = repo.Save(ctx, &message.Message{
			ID:          fmt.Sprintf("bmsg-%d-%s", i, uuid.New().String()[:8]),
			ChatID:      chatID,
			SenderID:    "bench-user",
			ContentType: "markdown",
			Content:     "Test content",
			Timestamp:   time.Now(),
		})
	}

	b.ResetTimer()
	for range b.N {
		if _, err := repo.FindByChatID(ctx, chatID, 50, time.Now()); err != nil {
			b.Fatal(err)
		}
	}
}
