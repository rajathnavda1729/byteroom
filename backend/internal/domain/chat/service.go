package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service contains business logic for chat management.
type Service struct {
	repo Repository
}

// NewService creates a new chat Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateGroup creates a new group chat, adding the creator as admin and optional
// initial members as regular members.
func (s *Service) CreateGroup(ctx context.Context, creatorID, name string, memberIDs []string) (*Chat, error) {
	allMembers := append([]string{creatorID}, memberIDs...)

	c := &Chat{
		ID:        uuid.New().String(),
		Name:      name,
		Type:      ChatTypeGroup,
		CreatedBy: creatorID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Members:   allMembers,
	}

	if c.Name == "" {
		return nil, ErrEmptyChatName
	}

	if err := s.repo.Save(ctx, c); err != nil {
		return nil, fmt.Errorf("saving chat: %w", err)
	}

	if err := s.repo.AddMember(ctx, c.ID, creatorID, RoleAdmin); err != nil {
		return nil, fmt.Errorf("adding creator as admin: %w", err)
	}

	for _, uid := range memberIDs {
		if err := s.repo.AddMember(ctx, c.ID, uid, RoleMember); err != nil {
			return nil, fmt.Errorf("adding member %s: %w", uid, err)
		}
	}

	return c, nil
}

// CreateDirect returns the existing direct chat between two users, or creates
// one if none exists. This ensures two users always share exactly one room.
func (s *Service) CreateDirect(ctx context.Context, userID1, userID2 string) (*Chat, error) {
	existing, err := s.repo.FindDirectBetween(ctx, userID1, userID2)
	if err == nil {
		return existing, nil
	}
	if err != ErrChatNotFound {
		return nil, fmt.Errorf("checking existing direct chat: %w", err)
	}

	c := &Chat{
		ID:        uuid.New().String(),
		Name:      "",
		Type:      ChatTypeDirect,
		CreatedBy: userID1,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Members:   []string{userID1, userID2},
	}

	if err := s.repo.Save(ctx, c); err != nil {
		return nil, fmt.Errorf("saving direct chat: %w", err)
	}

	for _, uid := range c.Members {
		if err := s.repo.AddMember(ctx, c.ID, uid, RoleMember); err != nil {
			return nil, fmt.Errorf("adding direct member: %w", err)
		}
	}

	return c, nil
}

// AddMember adds a new member to a chat. The requester must be an admin.
func (s *Service) AddMember(ctx context.Context, chatID, requesterID, newMemberID string) error {
	role, err := s.repo.GetMemberRole(ctx, chatID, requesterID)
	if err != nil {
		return ErrNotMember
	}
	if role != RoleAdmin {
		return ErrForbidden
	}

	if err := s.repo.AddMember(ctx, chatID, newMemberID, RoleMember); err != nil {
		return fmt.Errorf("adding member: %w", err)
	}

	return nil
}

// RemoveMember removes a member from a chat. Admins can remove anyone; members can
// only remove themselves (leave).
func (s *Service) RemoveMember(ctx context.Context, chatID, requesterID, targetID string) error {
	role, err := s.repo.GetMemberRole(ctx, chatID, requesterID)
	if err != nil {
		return ErrNotMember
	}

	if requesterID != targetID && role != RoleAdmin {
		return ErrForbidden
	}

	if err := s.repo.RemoveMember(ctx, chatID, targetID); err != nil {
		return fmt.Errorf("removing member: %w", err)
	}

	return nil
}

// GetByID retrieves a chat, verifying the requester is a member.
func (s *Service) GetByID(ctx context.Context, chatID, requesterID string) (*Chat, error) {
	c, err := s.repo.FindByID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("finding chat: %w", err)
	}

	isMember, err := s.repo.IsMember(ctx, chatID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("checking membership: %w", err)
	}
	if !isMember {
		return nil, ErrNotMember
	}

	return c, nil
}

// ListForUser returns all chats the user is a member of.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]*Chat, error) {
	chats, err := s.repo.FindByMember(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing chats: %w", err)
	}
	return chats, nil
}
