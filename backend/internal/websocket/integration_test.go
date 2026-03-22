//go:build integration

package websocket

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apihandler "github.com/byteroom/backend/internal/api/handler"
	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/byteroom/backend/internal/domain/chat"
	"github.com/byteroom/backend/internal/domain/message"
	"github.com/byteroom/backend/internal/domain/user"
	"github.com/byteroom/backend/internal/infrastructure/postgres"
	"github.com/byteroom/backend/internal/infrastructure/sanitizer"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupIntegrationDB connects to the test PostgreSQL instance defined in
// docker-compose.test.yml and runs the migrations.
func setupIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	cfg := &postgres.Config{
		Host:     "localhost",
		Port:     "5433",
		User:     "byteroom_test",
		Password: "byteroom_test",
		DBName:   "byteroom_test",
		SSLMode:  "disable",
	}
	db, err := postgres.Connect(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

type sanitizerAdapter struct{}

func (s sanitizerAdapter) Sanitize(content string) string {
	return sanitizer.SanitizeMarkdown(content)
}

func wsURL(server *httptest.Server, token string) string {
	return "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token
}

func createIntegrationUser(t *testing.T, svc *user.Service, username string) *user.User {
	t.Helper()
	u, err := svc.Register(context.Background(), username, "password123")
	require.NoError(t, err)
	return u
}

func TestWebSocket_FullMessageFlow(t *testing.T) {
	db := setupIntegrationDB(t)

	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)
	msgRepo := postgres.NewMessageRepository(db)

	// Services
	userSvc := user.NewService(userRepo)
	chatSvc := chat.NewService(chatRepo)
	msgSvc := message.NewService(msgRepo, chatRepo, sanitizerAdapter{})

	jwtMgr := middleware.NewJWTManager("integration-test-secret", 24*time.Hour)

	// Build WS event router
	wsRouter := NewEventRouter()
	msgHandler := NewMessageEventHandler(hub, msgSvc)
	typingHandler := NewTypingHandler(hub)

	wsRouter.RegisterHandler(EventMessageSend, msgHandler.HandleSend)
	wsRouter.RegisterHandler(EventTypingStart, func(_ context.Context, c *Client, f *WSFrame) error {
		typingHandler.HandleStart(c, f)
		return nil
	})
	wsRouter.RegisterHandler(EventTypingStop, func(_ context.Context, c *Client, f *WSFrame) error {
		typingHandler.HandleStop(c, f)
		return nil
	})
	wsRouter.RegisterHandler(EventPing, func(_ context.Context, c *Client, f *WSFrame) error {
		c.SendFrame(&WSFrame{Event: EventPong, Data: map[string]interface{}{}})
		return nil
	})

	// Create test users
	alice := createIntegrationUser(t, userSvc, fmt.Sprintf("alice_%d", time.Now().UnixNano()))
	bob := createIntegrationUser(t, userSvc, fmt.Sprintf("bob_%d", time.Now().UnixNano()))

	// Create a direct chat between alice and bob
	testChat, err := chatSvc.CreateDirect(context.Background(), alice.ID, bob.ID)
	require.NoError(t, err)

	// HTTP test server
	wsH := apihandler.NewWSHandler(hub, jwtMgr, chatRepo, wsRouter)
	server := httptest.NewServer(wsH)
	defer server.Close()

	// Connect alice
	tokenAlice, _ := jwtMgr.Generate(alice.ID)
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL(server, tokenAlice), nil)
	require.NoError(t, err)
	defer ws1.Close()

	// Connect bob
	tokenBob, _ := jwtMgr.Generate(bob.ID)
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL(server, tokenBob), nil)
	require.NoError(t, err)
	defer ws2.Close()

	// Give the hub time to process registrations
	time.Sleep(20 * time.Millisecond)

	// Alice sends a message
	msgID := fmt.Sprintf("msg-test-%d", time.Now().UnixNano())
	sendPayload := fmt.Sprintf(`{
		"event": "message.send",
		"request_id": "req-1",
		"data": {
			"message_id": %q,
			"chat_id": %q,
			"content_type": "markdown",
			"content": "Hello from Alice!"
		}
	}`, msgID, testChat.ID)
	err = ws1.WriteMessage(websocket.TextMessage, []byte(sendPayload))
	require.NoError(t, err)

	// Alice receives ACK
	ws1.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, ackBytes, err := ws1.ReadMessage()
	require.NoError(t, err)
	var ack WSFrame
	require.NoError(t, json.Unmarshal(ackBytes, &ack))
	assert.Equal(t, EventMessageAck, ack.Event)
	assert.Equal(t, "req-1", ack.RequestID)

	// Bob receives message.new
	ws2.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, newBytes, err := ws2.ReadMessage()
	require.NoError(t, err)
	var received WSFrame
	require.NoError(t, json.Unmarshal(newBytes, &received))
	assert.Equal(t, EventMessageNew, received.Event)
	assert.Equal(t, "Hello from Alice!", received.Data["content"])

	// Verify the message was persisted
	saved, err := msgRepo.FindByID(context.Background(), msgID)
	require.NoError(t, err)
	assert.Equal(t, "Hello from Alice!", saved.Content)
}

func TestWebSocket_TypingIndicators(t *testing.T) {
	db := setupIntegrationDB(t)

	hub := NewHub()
	go hub.Run()
	defer hub.Shutdown()

	userRepo := postgres.NewUserRepository(db)
	chatRepo := postgres.NewChatRepository(db)

	userSvc := user.NewService(userRepo)
	chatSvc := chat.NewService(chatRepo)
	jwtMgr := middleware.NewJWTManager("integration-test-secret", 24*time.Hour)

	wsRouter := NewEventRouter()
	typingHandler := NewTypingHandler(hub)

	wsRouter.RegisterHandler(EventTypingStart, func(_ context.Context, c *Client, f *WSFrame) error {
		typingHandler.HandleStart(c, f)
		return nil
	})
	wsRouter.RegisterHandler(EventTypingStop, func(_ context.Context, c *Client, f *WSFrame) error {
		typingHandler.HandleStop(c, f)
		return nil
	})

	alice := createIntegrationUser(t, userSvc, fmt.Sprintf("typing_alice_%d", time.Now().UnixNano()))
	bob := createIntegrationUser(t, userSvc, fmt.Sprintf("typing_bob_%d", time.Now().UnixNano()))
	testChat, err := chatSvc.CreateDirect(context.Background(), alice.ID, bob.ID)
	require.NoError(t, err)

	wsH := apihandler.NewWSHandler(hub, jwtMgr, chatRepo, wsRouter)
	server := httptest.NewServer(wsH)
	defer server.Close()

	tokenAlice, _ := jwtMgr.Generate(alice.ID)
	ws1, _, err := websocket.DefaultDialer.Dial(wsURL(server, tokenAlice), nil)
	require.NoError(t, err)
	defer ws1.Close()

	tokenBob, _ := jwtMgr.Generate(bob.ID)
	ws2, _, err := websocket.DefaultDialer.Dial(wsURL(server, tokenBob), nil)
	require.NoError(t, err)
	defer ws2.Close()

	time.Sleep(20 * time.Millisecond)

	// Alice starts typing
	typingStart := fmt.Sprintf(`{"event":"typing.start","data":{"chat_id":%q}}`, testChat.ID)
	err = ws1.WriteMessage(websocket.TextMessage, []byte(typingStart))
	require.NoError(t, err)

	// Bob receives user.typing (is_typing: true)
	ws2.SetReadDeadline(time.Now().Add(time.Second))
	_, typingBytes, err := ws2.ReadMessage()
	require.NoError(t, err)
	var typingFrame WSFrame
	require.NoError(t, json.Unmarshal(typingBytes, &typingFrame))
	assert.Equal(t, EventUserTyping, typingFrame.Event)
	assert.Equal(t, true, typingFrame.Data["is_typing"])
	assert.Equal(t, alice.ID, typingFrame.Data["user_id"])
}
