package message

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MembershipChecker verifies a user's membership in a chat.
type MembershipChecker interface {
	IsMember(ctx context.Context, chatID, userID string) (bool, error)
}

// Sanitizer cleans user-provided content.
type Sanitizer interface {
	Sanitize(content string) string
}

// Service contains business logic for message operations.
type Service struct {
	repo       Repository
	membership MembershipChecker
	sanitizer  Sanitizer
}

// NewService creates a new message Service.
func NewService(repo Repository, membership MembershipChecker, sanitizer Sanitizer) *Service {
	return &Service{repo: repo, membership: membership, sanitizer: sanitizer}
}

// Send persists a message. It verifies the sender is a chat member, sanitizes
// the content, and ensures idempotency via the message ID.
func (s *Service) Send(ctx context.Context, msg *Message) (*Message, error) {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	// Membership check
	if s.membership != nil {
		ok, err := s.membership.IsMember(ctx, msg.ChatID, msg.SenderID)
		if err != nil {
			return nil, fmt.Errorf("checking membership: %w", err)
		}
		if !ok {
			return nil, ErrForbidden
		}
	}

	// Sanitize content before validation so length check is on the clean text
	if s.sanitizer != nil {
		msg.Content = s.sanitizer.Sanitize(msg.Content)
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	// Idempotency check
	exists, err := s.repo.ExistsByID(ctx, msg.ID)
	if err != nil {
		return nil, fmt.Errorf("checking idempotency: %w", err)
	}
	if exists {
		return s.repo.FindByID(ctx, msg.ID)
	}

	msg.CreatedAt = time.Now().UTC()

	if err := s.repo.Save(ctx, msg); err != nil {
		return nil, fmt.Errorf("saving message: %w", err)
	}

	return msg, nil
}

// GetHistory retrieves paginated message history for a chat room.
func (s *Service) GetHistory(ctx context.Context, chatID string, limit, offset int) ([]*Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	msgs, err := s.repo.FindByChatID(ctx, chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("fetching history: %w", err)
	}

	return msgs, nil
}
