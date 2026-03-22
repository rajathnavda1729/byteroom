package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/byteroom/backend/internal/domain/chat"
	"github.com/byteroom/backend/internal/infrastructure/postgres"
)

// ChatService is the interface the ChatHandler depends on.
type ChatService interface {
	CreateGroup(ctx context.Context, creatorID, name string, memberIDs []string) (*chat.Chat, error)
	CreateDirect(ctx context.Context, userID1, userID2 string) (*chat.Chat, error)
	GetByID(ctx context.Context, chatID, requesterID string) (*chat.Chat, error)
	ListForUser(ctx context.Context, userID string) ([]*chat.Chat, error)
	AddMember(ctx context.Context, chatID, requesterID, newMemberID string) error
	RemoveMember(ctx context.Context, chatID, requesterID, targetID string) error
}

// ChatHandler handles chat room endpoints.
type ChatHandler struct {
	chats    ChatService
	members  MemberDetailer
	notifier ChatNotifier
}

// NewChatHandler creates a ChatHandler.
func NewChatHandler(chats ChatService, members MemberDetailer) *ChatHandler {
	return &ChatHandler{chats: chats, members: members}
}

// WithNotifier attaches a real-time notifier (the WS hub) for post-create events.
func (h *ChatHandler) WithNotifier(n ChatNotifier) *ChatHandler {
	h.notifier = n
	return h
}

type createChatRequest struct {
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	MemberIDs []string `json:"member_ids"`
}

type addMemberRequest struct {
	UserID string `json:"user_id"`
}

type chatMemberDTO struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type chatDTO struct {
	ChatID    string          `json:"chat_id"`
	Name      string          `json:"name"`
	Type      string          `json:"type"`
	Members   []chatMemberDTO `json:"members"`
	CreatedBy string          `json:"created_by"`
	CreatedAt string          `json:"created_at"`
}

// MemberDetailer is the minimal DB interface needed to look up member display names.
type MemberDetailer interface {
	GetMemberDetails(ctx context.Context, chatID string) ([]postgres.MemberDetail, error)
}

// ChatNotifier handles real-time notifications when chats are created.
type ChatNotifier interface {
	SubscribeUserToRoom(userID, chatID string)
	BroadcastToUser(userID string, data []byte)
}

// List handles GET /api/chats.
func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	chats, err := h.chats.ListForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list chats")
		return
	}

	dtos := make([]chatDTO, 0, len(chats))
	for _, c := range chats {
		dto := toChatDTO(c)
		if h.members != nil {
			details, err := h.members.GetMemberDetails(r.Context(), c.ID)
			if err == nil {
				dto.Members = toMemberDTOs(details)
			}
		}
		dtos = append(dtos, dto)
	}
	writeJSON(w, http.StatusOK, map[string]any{"chats": dtos})
}

// Create handles POST /api/chats.
func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req createChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var (
		c   *chat.Chat
		err error
	)

	switch req.Type {
	case "group":
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "name is required for group chats")
			return
		}
		c, err = h.chats.CreateGroup(r.Context(), userID, req.Name, req.MemberIDs)
	case "direct":
		if len(req.MemberIDs) != 1 {
			writeError(w, http.StatusBadRequest, "direct chat requires exactly one other member_id")
			return
		}
		c, err = h.chats.CreateDirect(r.Context(), userID, req.MemberIDs[0])
	default:
		writeError(w, http.StatusBadRequest, "type must be 'group' or 'direct'")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create chat")
		return
	}

	// Enrich the response with member names
	dto := toChatDTO(c)
	if h.members != nil {
		if details, detailErr := h.members.GetMemberDetails(r.Context(), c.ID); detailErr == nil {
			dto.Members = toMemberDTOs(details)
		}
	}

	// Notify all members via WebSocket: subscribe them to the room and push chat.new
	if h.notifier != nil {
		notifyData := buildChatNewFrame(dto)
		for _, m := range dto.Members {
			h.notifier.SubscribeUserToRoom(m.UserID, c.ID)
			h.notifier.BroadcastToUser(m.UserID, notifyData)
		}
	}

	writeJSON(w, http.StatusCreated, dto)
}

// GetByID handles GET /api/chats/{id}.
func (h *ChatHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	chatID := r.PathValue("id")
	c, err := h.chats.GetByID(r.Context(), chatID, userID)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrChatNotFound):
			writeError(w, http.StatusNotFound, "chat not found")
		case errors.Is(err, chat.ErrNotMember):
			writeError(w, http.StatusForbidden, "access denied")
		default:
			writeError(w, http.StatusInternalServerError, "failed to get chat")
		}
		return
	}

	writeJSON(w, http.StatusOK, toChatDTO(c))
}

// AddMember handles POST /api/chats/{id}/members.
func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	chatID := r.PathValue("id")

	var req addMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if err := h.chats.AddMember(r.Context(), chatID, userID, req.UserID); err != nil {
		switch {
		case errors.Is(err, chat.ErrForbidden):
			writeError(w, http.StatusForbidden, "only admins can add members")
		case errors.Is(err, chat.ErrNotMember):
			writeError(w, http.StatusForbidden, "access denied")
		default:
			writeError(w, http.StatusInternalServerError, "failed to add member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember handles DELETE /api/chats/{id}/members/{userId}.
func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	chatID := r.PathValue("id")
	targetID := r.PathValue("userId")

	if err := h.chats.RemoveMember(r.Context(), chatID, userID, targetID); err != nil {
		switch {
		case errors.Is(err, chat.ErrForbidden):
			writeError(w, http.StatusForbidden, "only admins can remove other members")
		case errors.Is(err, chat.ErrNotMember):
			writeError(w, http.StatusForbidden, "access denied")
		default:
			writeError(w, http.StatusInternalServerError, "failed to remove member")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toChatDTO(c *chat.Chat) chatDTO {
	return chatDTO{
		ChatID:    c.ID,
		Name:      c.Name,
		Type:      string(c.Type),
		Members:   []chatMemberDTO{},
		CreatedBy: c.CreatedBy,
		CreatedAt: c.CreatedAt.String(),
	}
}

func buildChatNewFrame(dto chatDTO) []byte {
	frame := map[string]any{
		"event": "chat.new",
		"data":  dto,
	}
	b, err := json.Marshal(frame)
	if err != nil {
		log.Printf("chat_handler: marshal chat.new: %v", err)
		return nil
	}
	return b
}

func toMemberDTOs(details []postgres.MemberDetail) []chatMemberDTO {
	dtos := make([]chatMemberDTO, 0, len(details))
	for _, d := range details {
		dtos = append(dtos, chatMemberDTO{
			UserID:      d.UserID,
			Username:    d.Username,
			DisplayName: d.DisplayName,
			AvatarURL:   d.AvatarURL,
		})
	}
	return dtos
}
