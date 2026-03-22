package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/byteroom/backend/internal/domain/message"
	ws "github.com/byteroom/backend/internal/websocket"
)

// MessageService is the interface the MessageHandler depends on.
type MessageService interface {
	GetHistory(ctx context.Context, chatID string, limit, offset int) ([]*message.Message, error)
	Send(ctx context.Context, msg *message.Message) (*message.Message, error)
}

// MessageHub broadcasts persisted messages to WebSocket subscribers.
type MessageHub interface {
	BroadcastToChatExceptUser(chatID string, data []byte, excludeUserID string)
}

// MessageHandler handles message endpoints.
type MessageHandler struct {
	msgs MessageService
	hub  MessageHub
}

// NewMessageHandler creates a MessageHandler.
func NewMessageHandler(msgs MessageService) *MessageHandler {
	return &MessageHandler{msgs: msgs}
}

// WithHub attaches the WebSocket hub so POST /messages can notify other clients.
func (h *MessageHandler) WithHub(hub MessageHub) *MessageHandler {
	h.hub = hub
	return h
}

type sendMessageRequest struct {
	MessageID   string `json:"message_id"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type messageDTO struct {
	MessageID   string `json:"message_id"`
	ChatID      string `json:"chat_id"`
	SenderID    string `json:"sender_id"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
	Timestamp   string `json:"timestamp"`
}

// GetHistory handles GET /api/chats/{id}/messages.
func (h *MessageHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	chatID := r.PathValue("id")

	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	msgs, err := h.msgs.GetHistory(r.Context(), chatID, limit, offset)
	if err != nil {
		if errors.Is(err, message.ErrForbidden) {
			writeError(w, http.StatusForbidden, "access denied")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get messages")
		return
	}

	dtos := make([]messageDTO, 0, len(msgs))
	for _, m := range msgs {
		dtos = append(dtos, toMessageDTO(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"messages": dtos})
}

// Send handles POST /api/chats/{id}/messages — persists a message and broadcasts
// to WebSocket room members (same path as WS message.send, but reliable over HTTP).
func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	chatID := r.PathValue("id")
	if chatID == "" {
		writeError(w, http.StatusBadRequest, "missing chat id")
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ContentType == "" {
		req.ContentType = string(message.ContentTypeMarkdown)
	}

	msg := &message.Message{
		ID:          req.MessageID,
		ChatID:      chatID,
		SenderID:    userID,
		ContentType: message.ContentType(req.ContentType),
		Content:     req.Content,
		CreatedAt:   time.Now().UTC(),
	}

	saved, err := h.msgs.Send(r.Context(), msg)
	if err != nil {
		switch {
		case errors.Is(err, message.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden")
		case errors.Is(err, message.ErrInvalidContent), errors.Is(err, message.ErrContentTooLong):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to send message")
		}
		return
	}

	if h.hub != nil {
		frame := map[string]any{
			"event": ws.EventMessageNew,
			"data": map[string]any{
				"message_id":   saved.ID,
				"chat_id":      saved.ChatID,
				"sender_id":    saved.SenderID,
				"content_type": string(saved.ContentType),
				"content":      saved.Content,
				"created_at":   saved.CreatedAt.UTC().Format(time.RFC3339),
				"timestamp":    saved.CreatedAt.UTC().Format(time.RFC3339),
			},
		}
		if raw, err := json.Marshal(frame); err == nil {
			h.hub.BroadcastToChatExceptUser(chatID, raw, userID)
		}
	}

	writeJSON(w, http.StatusCreated, toMessageDTO(saved))
}

func toMessageDTO(m *message.Message) messageDTO {
	return messageDTO{
		MessageID:   m.ID,
		ChatID:      m.ChatID,
		SenderID:    m.SenderID,
		ContentType: string(m.ContentType),
		Content:     m.Content,
		Timestamp:   m.CreatedAt.UTC().Format(time.RFC3339),
	}
}
