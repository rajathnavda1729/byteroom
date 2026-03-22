package handler

import (
	"context"
	"log"
	"net/http"

	ws "github.com/byteroom/backend/internal/websocket"
	"github.com/gorilla/websocket"
)

// TokenValidator is the subset of middleware.JWTManager used by the WS handler.
type TokenValidator interface {
	Validate(token string) (string, error)
}

// ChatMemberLister fetches all chat IDs a user belongs to.
type ChatMemberLister interface {
	GetUserChatIDs(ctx context.Context, userID string) ([]string, error)
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	// In production this should validate the Origin header.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler upgrades HTTP connections to WebSocket after validating the JWT
// supplied in the `?token=` query parameter.
type WSHandler struct {
	hub    *ws.Hub
	jwt    TokenValidator
	chatML ChatMemberLister
	router ws.EventRouter
}

// NewWSHandler constructs a WSHandler. router may be nil during tests that
// only check authentication behaviour.
func NewWSHandler(hub *ws.Hub, jwt TokenValidator, chatML ChatMemberLister, router ws.EventRouter) *WSHandler {
	return &WSHandler{hub: hub, jwt: jwt, chatML: chatML, router: router}
}

// ServeHTTP validates the JWT, fetches the user's chat memberships, upgrades
// the connection, and starts the read/write pumps.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	userID, err := h.jwt.Validate(token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	var chatIDs []string
	if h.chatML != nil {
		chatIDs, err = h.chatML.GetUserChatIDs(r.Context(), userID)
		if err != nil {
			log.Printf("ws: failed to fetch chat ids for user %s: %v", userID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade already wrote the error response.
		return
	}

	client := ws.NewClient(h.hub, conn, userID, chatIDs, h.router)

	// Register synchronously with the hub before starting pumps. Otherwise an
	// immediate HTTP message.send broadcast can be processed before this client
	// is added to chat rooms (multi-channel select order is not FIFO).
	registered := make(chan struct{})
	h.hub.RegisterWithAck(client, registered)
	<-registered

	go client.WritePump()
	go client.ReadPump()
}
