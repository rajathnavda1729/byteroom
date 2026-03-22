package chat

import "context"

// Repository defines persistence operations for chats.
type Repository interface {
	Save(ctx context.Context, c *Chat) error
	FindByID(ctx context.Context, id string) (*Chat, error)
	FindByMember(ctx context.Context, userID string) ([]*Chat, error)
	// FindDirectBetween returns the existing direct chat between two users, or
	// ErrChatNotFound if no such chat exists.
	FindDirectBetween(ctx context.Context, userID1, userID2 string) (*Chat, error)
	GetUserChatIDs(ctx context.Context, userID string) ([]string, error)
	AddMember(ctx context.Context, chatID, userID string, role MemberRole) error
	RemoveMember(ctx context.Context, chatID, userID string) error
	GetMemberRole(ctx context.Context, chatID, userID string) (MemberRole, error)
	IsMember(ctx context.Context, chatID, userID string) (bool, error)
	Delete(ctx context.Context, id string) error
}
