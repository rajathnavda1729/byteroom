package message

import "context"

// Repository defines persistence operations for messages.
type Repository interface {
	Save(ctx context.Context, msg *Message) error
	FindByID(ctx context.Context, id string) (*Message, error)
	FindByChatID(ctx context.Context, chatID string, limit, offset int) ([]*Message, error)
	ExistsByID(ctx context.Context, id string) (bool, error)
}
