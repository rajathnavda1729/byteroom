package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const defaultTypingTimeout = 5 * time.Second

// TypingHandler manages typing indicator state and broadcasts user.typing
// events to chat members. It auto-stops indicators after a configurable
// timeout (default 5 s) with no explicit stop from the client.
type TypingHandler struct {
	hub     *Hub
	timeout time.Duration

	mu     sync.Mutex
	timers map[string]*time.Timer // key: "userID:chatID"
}

// NewTypingHandler creates a TypingHandler with the default 5-second timeout.
func NewTypingHandler(hub *Hub) *TypingHandler {
	return &TypingHandler{
		hub:     hub,
		timeout: defaultTypingTimeout,
		timers:  make(map[string]*time.Timer),
	}
}

// SetTimeout overrides the auto-stop timeout (used in tests).
func (h *TypingHandler) SetTimeout(d time.Duration) {
	h.timeout = d
}

// HandleStart broadcasts a typing.start indicator and schedules an auto-stop.
func (h *TypingHandler) HandleStart(sender *Client, f *WSFrame) {
	chatID, _ := f.Data["chat_id"].(string)
	key := timerKey(sender.userID, chatID)

	h.mu.Lock()
	// Reset any existing timer for this user/chat combination.
	if t, ok := h.timers[key]; ok {
		t.Stop()
	}
	h.timers[key] = time.AfterFunc(h.timeout, func() {
		h.broadcastTyping(sender, chatID, false)
		h.mu.Lock()
		delete(h.timers, key)
		h.mu.Unlock()
	})
	h.mu.Unlock()

	h.broadcastTyping(sender, chatID, true)
}

// HandleStop broadcasts a typing.stop indicator and cancels any pending
// auto-stop timer.
func (h *TypingHandler) HandleStop(sender *Client, f *WSFrame) {
	chatID, _ := f.Data["chat_id"].(string)
	key := timerKey(sender.userID, chatID)

	h.mu.Lock()
	if t, ok := h.timers[key]; ok {
		t.Stop()
		delete(h.timers, key)
	}
	h.mu.Unlock()

	h.broadcastTyping(sender, chatID, false)
}

// broadcastTyping marshals and sends a user.typing frame to all chat members
// except the sender.
func (h *TypingHandler) broadcastTyping(sender *Client, chatID string, isTyping bool) {
	frame := &WSFrame{
		Event: EventUserTyping,
		Data: map[string]interface{}{
			"chat_id":   chatID,
			"user_id":   sender.userID,
			"is_typing": isTyping,
		},
	}
	data, err := json.Marshal(frame)
	if err != nil {
		return
	}
	h.hub.BroadcastToChat(chatID, data, sender)
}

func timerKey(userID, chatID string) string {
	return fmt.Sprintf("%s:%s", userID, chatID)
}
