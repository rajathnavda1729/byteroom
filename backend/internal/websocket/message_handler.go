package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/byteroom/backend/internal/domain/message"
)

// MessageService is the subset of message.Service used by the WS handler.
type MessageService interface {
	Send(ctx context.Context, msg *message.Message) (*message.Message, error)
}

// MessageEventHandler handles the "message.send" WebSocket event.
type MessageEventHandler struct {
	hub    *Hub
	msgSvc MessageService
}

// NewMessageEventHandler constructs a MessageEventHandler.
func NewMessageEventHandler(hub *Hub, msgSvc MessageService) *MessageEventHandler {
	return &MessageEventHandler{hub: hub, msgSvc: msgSvc}
}

// HandleSend processes an incoming message.send frame:
//  1. Persist the message via MessageService.
//  2. Send an ACK frame back to the sender.
//  3. Broadcast a message.new frame to all other chat members.
func (h *MessageEventHandler) HandleSend(ctx context.Context, sender *Client, f *WSFrame) error {
	chatID, _ := f.Data["chat_id"].(string)
	content, _ := f.Data["content"].(string)
	contentType, _ := f.Data["content_type"].(string)
	messageID, _ := f.Data["message_id"].(string)

	if contentType == "" {
		contentType = string(message.ContentTypeMarkdown)
	}

	msg := &message.Message{
		ID:          messageID,
		ChatID:      chatID,
		SenderID:    sender.userID,
		ContentType: message.ContentType(contentType),
		Content:     content,
		CreatedAt:   time.Now().UTC(),
	}

	saved, err := h.msgSvc.Send(ctx, msg)
	if err != nil {
		h.sendError(sender, f.RequestID, err)
		return err
	}

	// ACK to sender
	sender.SendFrame(&WSFrame{
		Event:     EventMessageAck,
		RequestID: f.RequestID,
		Data: map[string]interface{}{
			"message_id": saved.ID,
			"status":     true,
		},
	})

	// Broadcast to all other members of the chat
	broadcast := &WSFrame{
		Event: EventMessageNew,
		Data: map[string]interface{}{
			"message_id":   saved.ID,
			"chat_id":      saved.ChatID,
			"sender_id":    saved.SenderID,
			"content_type": string(saved.ContentType),
			"content":      saved.Content,
			"created_at":   saved.CreatedAt.Format(time.RFC3339),
			"timestamp":    saved.CreatedAt.Format(time.RFC3339),
		},
	}
	data, err := json.Marshal(broadcast)
	if err != nil {
		return fmt.Errorf("marshalling broadcast: %w", err)
	}
	h.hub.BroadcastToChat(chatID, data, sender)

	return nil
}

func (h *MessageEventHandler) sendError(c *Client, requestID string, err error) {
	msg := err.Error()
	if errors.Is(err, message.ErrForbidden) {
		msg = "forbidden"
	}
	c.SendFrame(&WSFrame{
		Event:     EventMessageError,
		RequestID: requestID,
		Data:      map[string]interface{}{"error": msg},
	})
}
