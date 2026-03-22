package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockMessageService struct{ mock.Mock }

func (m *MockMessageService) GetHistory(ctx context.Context, chatID string, limit, offset int) ([]*message.Message, error) {
	args := m.Called(ctx, chatID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*message.Message), args.Error(1)
}

func (m *MockMessageService) Send(ctx context.Context, msg *message.Message) (*message.Message, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*message.Message), args.Error(1)
}

func makeMessage(id, chatID string) *message.Message {
	return &message.Message{
		ID:          id,
		ChatID:      chatID,
		SenderID:    "user-1",
		ContentType: message.ContentTypeMarkdown,
		Content:     "Hello world",
		CreatedAt:   time.Now(),
	}
}

func TestMessageHandler_GetHistory_Returns200WithMessages(t *testing.T) {
	svc := new(MockMessageService)
	h := NewMessageHandler(svc)

	msgs := []*message.Message{makeMessage("msg-1", "chat-1"), makeMessage("msg-2", "chat-1")}
	svc.On("GetHistory", mock.Anything, "chat-1", 50, 0).Return(msgs, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/chats/chat-1/messages", nil)
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetHistory(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result struct {
		Messages []messageDTO `json:"messages"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Len(t, result.Messages, 2)
}

func TestMessageHandler_GetHistory_RespectsLimitAndOffset(t *testing.T) {
	svc := new(MockMessageService)
	h := NewMessageHandler(svc)

	svc.On("GetHistory", mock.Anything, "chat-1", 20, 40).Return([]*message.Message{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/chats/chat-1/messages?limit=20&offset=40", nil)
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetHistory(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	svc.AssertExpectations(t)
}

func TestMessageHandler_GetHistory_NotAuthenticated_Returns401(t *testing.T) {
	h := NewMessageHandler(new(MockMessageService))

	req := httptest.NewRequest(http.MethodGet, "/api/chats/chat-1/messages", nil)
	req.SetPathValue("id", "chat-1")
	rec := httptest.NewRecorder()

	h.GetHistory(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestMessageHandler_GetHistory_EmptyResult_ReturnsEmptyArray(t *testing.T) {
	svc := new(MockMessageService)
	h := NewMessageHandler(svc)

	svc.On("GetHistory", mock.Anything, "chat-1", 50, 0).Return([]*message.Message{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/chats/chat-1/messages", nil)
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetHistory(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result struct {
		Messages []messageDTO `json:"messages"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Len(t, result.Messages, 0)
}

type mockMessageHub struct{ mock.Mock }

func (m *mockMessageHub) BroadcastToChatExceptUser(chatID string, data []byte, excludeUserID string) {
	m.Called(chatID, data, excludeUserID)
}

func TestMessageHandler_Send_ValidBody_Returns201(t *testing.T) {
	svc := new(MockMessageService)
	hub := new(mockMessageHub)
	h := NewMessageHandler(svc).WithHub(hub)

	saved := makeMessage("msg-new", "chat-1")
	svc.On("Send", mock.Anything, mock.MatchedBy(func(m *message.Message) bool {
		return m.ChatID == "chat-1" && m.SenderID == "user-1" && m.Content == "hi"
	})).Return(saved, nil)
	hub.On("BroadcastToChatExceptUser", "chat-1", mock.Anything, "user-1").Return()

	body := `{"message_id":"msg-new","content":"hi","content_type":"markdown"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats/chat-1/messages", strings.NewReader(body))
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Send(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var dto messageDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
	assert.Equal(t, "msg-new", dto.MessageID)
	svc.AssertExpectations(t)
	hub.AssertExpectations(t)
}
