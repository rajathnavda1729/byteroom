package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ws "github.com/byteroom/backend/internal/websocket"
	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockChatMemberLister satisfies ChatMemberLister.
type MockChatMemberLister struct {
	mock.Mock
}

func (m *MockChatMemberLister) GetUserChatIDs(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestWSHandler_Upgrade_ValidToken_Connects(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	jwtManager := middleware.NewJWTManager("secret", 24*time.Hour)

	chatRepo := new(MockChatMemberLister)
	chatRepo.On("GetUserChatIDs", mock.Anything, "user-123").
		Return([]string{"chat-1", "chat-2"}, nil)

	handler := NewWSHandler(hub, jwtManager, chatRepo, nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, err := jwtManager.Generate("user-123")
	assert.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	assert.NotNil(t, conn)
	conn.Close()
}

func TestWSHandler_Upgrade_InvalidToken_Returns401(t *testing.T) {
	hub := ws.NewHub()
	jwtManager := middleware.NewJWTManager("secret", 24*time.Hour)
	handler := NewWSHandler(hub, jwtManager, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ws?token=invalid", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWSHandler_Upgrade_MissingToken_Returns401(t *testing.T) {
	hub := ws.NewHub()
	jwtManager := middleware.NewJWTManager("secret", 24*time.Hour)
	handler := NewWSHandler(hub, jwtManager, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWSHandler_Upgrade_ExpiredToken_Returns401(t *testing.T) {
	hub := ws.NewHub()
	jwtManager := middleware.NewJWTManager("secret", -1*time.Hour) // already expired
	handler := NewWSHandler(hub, jwtManager, nil, nil)

	token, err := jwtManager.Generate("user-123")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/ws?token="+token, nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWSHandler_Upgrade_ChatRepoError_Returns500(t *testing.T) {
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Shutdown()

	jwtManager := middleware.NewJWTManager("secret", 24*time.Hour)

	chatRepo := new(MockChatMemberLister)
	chatRepo.On("GetUserChatIDs", mock.Anything, "user-123").
		Return(nil, assert.AnError)

	handler := NewWSHandler(hub, jwtManager, chatRepo, nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	token, _ := jwtManager.Generate("user-123")
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?token=" + token

	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
