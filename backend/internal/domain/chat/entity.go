package chat

import (
	"errors"
	"time"
)

var (
	ErrChatNotFound    = errors.New("chat not found")
	ErrInvalidChatType = errors.New("chat type must be 'direct' or 'group'")
	ErrEmptyChatName   = errors.New("group chat name cannot be empty")
	ErrNotMember       = errors.New("user is not a member of this chat")
	ErrForbidden       = errors.New("only admins can perform this action")
	ErrDirectChatSize  = errors.New("direct chat must have exactly 2 members")
)

// ChatType represents the kind of chat room.
type ChatType string

const (
	ChatTypeDirect ChatType = "direct"
	ChatTypeGroup  ChatType = "group"
)

// MemberRole defines a member's permissions within a chat.
type MemberRole string

const (
	RoleAdmin  MemberRole = "admin"
	RoleMember MemberRole = "member"
)

// Chat represents a chat room (direct or group).
type Chat struct {
	ID        string    `json:"chat_id"    db:"id"`
	Name      string    `json:"name"       db:"name"`
	Type      ChatType  `json:"type"       db:"type"`
	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Members   []string  `json:"members"    db:"-"`
}

// ChatMember records a user's membership and role in a chat.
type ChatMember struct {
	ChatID   string     `db:"chat_id"`
	UserID   string     `db:"user_id"`
	Role     MemberRole `db:"role"`
	JoinedAt time.Time  `db:"joined_at"`
}

// Validate ensures the chat fields are consistent.
func (c *Chat) Validate() error {
	if c.Type != ChatTypeDirect && c.Type != ChatTypeGroup {
		return ErrInvalidChatType
	}
	if c.Type == ChatTypeGroup && c.Name == "" {
		return ErrEmptyChatName
	}
	if c.Type == ChatTypeDirect && len(c.Members) != 2 {
		return ErrDirectChatSize
	}
	return nil
}

// IsMember checks whether userID is in the Members slice.
func (c *Chat) IsMember(userID string) bool {
	for _, m := range c.Members {
		if m == userID {
			return true
		}
	}
	return false
}
