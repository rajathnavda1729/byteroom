package message

import (
	"errors"
	"time"
)

var (
	ErrMessageNotFound    = errors.New("message not found")
	ErrDuplicateMessageID = errors.New("duplicate message ID")
	ErrInvalidContent     = errors.New("invalid message content")
	ErrContentTooLong     = errors.New("content exceeds maximum length of 50000 characters")
	ErrForbidden          = errors.New("user is not a member of this chat")
)

// ContentType enumerates allowed message payload formats.
type ContentType string

const (
	ContentTypeMarkdown     ContentType = "markdown"
	ContentTypeDiagramState ContentType = "diagram_state"
	ContentTypeImage        ContentType = "image"

	MaxContentLength = 50_000
)

// Message represents a chat message in the system.
type Message struct {
	ID          string      `json:"message_id"    db:"id"`
	ChatID      string      `json:"chat_id"       db:"chat_id"`
	SenderID    string      `json:"sender_id"     db:"sender_id"`
	ContentType ContentType `json:"content_type"  db:"content_type"`
	Content     string      `json:"content"       db:"content"`
	CreatedAt   time.Time   `json:"timestamp"     db:"created_at"`
}

// Validate ensures the message has well-formed fields.
func (m *Message) Validate() error {
	if m.Content == "" {
		return ErrInvalidContent
	}
	if len(m.Content) > MaxContentLength {
		return ErrContentTooLong
	}
	if m.ChatID == "" || m.SenderID == "" {
		return ErrInvalidContent
	}
	return nil
}
