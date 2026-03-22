package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/byteroom/backend/internal/domain/message"
)

// MessageRepository is a PostgreSQL implementation of message.Repository.
type MessageRepository struct {
	db *DB
}

// NewMessageRepository creates a new MessageRepository.
func NewMessageRepository(db *DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Save(ctx context.Context, msg *message.Message) error {
	// ON CONFLICT DO NOTHING ensures idempotency by message ID
	const q = `
		INSERT INTO messages (id, chat_id, sender_id, content_type, content, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, q,
		msg.ID, msg.ChatID, msg.SenderID, string(msg.ContentType), msg.Content, msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("saving message: %w", err)
	}
	return nil
}

func (r *MessageRepository) FindByID(ctx context.Context, id string) (*message.Message, error) {
	const q = `
		SELECT id, chat_id, sender_id, content_type, content, created_at
		FROM messages WHERE id = $1`

	msg := &message.Message{}
	var ct string
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&msg.ID, &msg.ChatID, &msg.SenderID, &ct, &msg.Content, &msg.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, message.ErrMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding message by id: %w", err)
	}
	msg.ContentType = message.ContentType(ct)
	return msg, nil
}

// FindByChatID returns up to `limit` messages for a chat starting at `offset`,
// ordered by created_at descending (newest first).
func (r *MessageRepository) FindByChatID(ctx context.Context, chatID string, limit, offset int) ([]*message.Message, error) {
	const q = `
		SELECT id, chat_id, sender_id, content_type, content, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, q, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("querying messages: %w", err)
	}
	defer rows.Close()

	var msgs []*message.Message
	for rows.Next() {
		msg := &message.Message{}
		var ct string
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &ct, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning message row: %w", err)
		}
		msg.ContentType = message.ContentType(ct)
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func (r *MessageRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	const q = `SELECT 1 FROM messages WHERE id = $1 LIMIT 1`

	var dummy int
	err := r.db.QueryRowContext(ctx, q, id).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking message exists: %w", err)
	}
	return true, nil
}
