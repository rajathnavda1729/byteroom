package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/byteroom/backend/internal/domain/chat"
)

// ChatRepository is a PostgreSQL implementation of chat.Repository.
type ChatRepository struct {
	db *DB
}

// NewChatRepository creates a new ChatRepository.
func NewChatRepository(db *DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Save(ctx context.Context, c *chat.Chat) error {
	const q = `
		INSERT INTO chats (id, name, type, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, q,
		c.ID, c.Name, string(c.Type), c.CreatedBy, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("saving chat: %w", err)
	}
	return nil
}

func (r *ChatRepository) FindByID(ctx context.Context, id string) (*chat.Chat, error) {
	const q = `
		SELECT id, name, type, created_by, created_at, updated_at
		FROM chats WHERE id = $1`

	c := &chat.Chat{}
	var chatType string
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&c.ID, &c.Name, &chatType, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, chat.ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding chat by id: %w", err)
	}
	c.Type = chat.ChatType(chatType)

	members, err := r.getMemberIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Members = members

	return c, nil
}

// FindDirectBetween returns the existing direct chat room shared by exactly
// the two given users, or chat.ErrChatNotFound if none exists.
func (r *ChatRepository) FindDirectBetween(ctx context.Context, userID1, userID2 string) (*chat.Chat, error) {
	const q = `
		SELECT c.id, c.name, c.type, c.created_by, c.created_at, c.updated_at
		FROM chats c
		WHERE c.type = 'direct'
		  AND EXISTS (SELECT 1 FROM chat_members WHERE chat_id = c.id AND user_id = $1)
		  AND EXISTS (SELECT 1 FROM chat_members WHERE chat_id = c.id AND user_id = $2)
		LIMIT 1`

	c := &chat.Chat{}
	var chatType string
	err := r.db.QueryRowContext(ctx, q, userID1, userID2).Scan(
		&c.ID, &c.Name, &chatType, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, chat.ErrChatNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("finding direct chat between users: %w", err)
	}
	c.Type = chat.ChatType(chatType)
	return c, nil
}

func (r *ChatRepository) FindByMember(ctx context.Context, userID string) ([]*chat.Chat, error) {
	const q = `
		SELECT c.id, c.name, c.type, c.created_by, c.created_at, c.updated_at
		FROM chats c
		INNER JOIN chat_members cm ON cm.chat_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC`

	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("finding chats by member: %w", err)
	}
	defer rows.Close()

	var chats []*chat.Chat
	for rows.Next() {
		c := &chat.Chat{}
		var chatType string
		if err := rows.Scan(&c.ID, &c.Name, &chatType, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning chat row: %w", err)
		}
		c.Type = chat.ChatType(chatType)
		chats = append(chats, c)
	}
	return chats, rows.Err()
}

func (r *ChatRepository) GetUserChatIDs(ctx context.Context, userID string) ([]string, error) {
	const q = `SELECT chat_id FROM chat_members WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user chat ids: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning chat id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *ChatRepository) AddMember(ctx context.Context, chatID, userID string, role chat.MemberRole) error {
	const q = `
		INSERT INTO chat_members (chat_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id, user_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, q, chatID, userID, string(role))
	if err != nil {
		return fmt.Errorf("adding member: %w", err)
	}
	return nil
}

func (r *ChatRepository) RemoveMember(ctx context.Context, chatID, userID string) error {
	const q = `DELETE FROM chat_members WHERE chat_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, q, chatID, userID)
	if err != nil {
		return fmt.Errorf("removing member: %w", err)
	}
	return nil
}

func (r *ChatRepository) GetMemberRole(ctx context.Context, chatID, userID string) (chat.MemberRole, error) {
	const q = `SELECT role FROM chat_members WHERE chat_id = $1 AND user_id = $2`

	var role string
	err := r.db.QueryRowContext(ctx, q, chatID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", chat.ErrNotMember
	}
	if err != nil {
		return "", fmt.Errorf("getting member role: %w", err)
	}
	return chat.MemberRole(role), nil
}

func (r *ChatRepository) IsMember(ctx context.Context, chatID, userID string) (bool, error) {
	const q = `SELECT 1 FROM chat_members WHERE chat_id = $1 AND user_id = $2 LIMIT 1`

	var dummy int
	err := r.db.QueryRowContext(ctx, q, chatID, userID).Scan(&dummy)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("checking membership: %w", err)
	}
	return true, nil
}

func (r *ChatRepository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM chats WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("deleting chat: %w", err)
	}
	return nil
}

func (r *ChatRepository) getMemberIDs(ctx context.Context, chatID string) ([]string, error) {
	const q = `SELECT user_id FROM chat_members WHERE chat_id = $1`
	rows, err := r.db.QueryContext(ctx, q, chatID)
	if err != nil {
		return nil, fmt.Errorf("getting members: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scanning member id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// MemberDetail holds the display information for a chat member.
type MemberDetail struct {
	UserID      string
	Username    string
	DisplayName string
	AvatarURL   string
}

// GetMemberDetails returns rich member info (joined with users table) for chatID.
func (r *ChatRepository) GetMemberDetails(ctx context.Context, chatID string) ([]MemberDetail, error) {
	const q = `
		SELECT u.id, u.username, u.display_name, COALESCE(u.avatar_url, '')
		FROM users u
		INNER JOIN chat_members cm ON cm.user_id = u.id
		WHERE cm.chat_id = $1
		ORDER BY u.username`

	rows, err := r.db.QueryContext(ctx, q, chatID)
	if err != nil {
		return nil, fmt.Errorf("getting member details: %w", err)
	}
	defer rows.Close()

	var members []MemberDetail
	for rows.Next() {
		var m MemberDetail
		if err := rows.Scan(&m.UserID, &m.Username, &m.DisplayName, &m.AvatarURL); err != nil {
			return nil, fmt.Errorf("scanning member detail: %w", err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
