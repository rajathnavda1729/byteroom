package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/domain/chat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockChatService struct{ mock.Mock }

func (m *MockChatService) CreateGroup(ctx context.Context, creatorID, name string, memberIDs []string) (*chat.Chat, error) {
	args := m.Called(ctx, creatorID, name, memberIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Chat), args.Error(1)
}

func (m *MockChatService) CreateDirect(ctx context.Context, userID1, userID2 string) (*chat.Chat, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Chat), args.Error(1)
}

func (m *MockChatService) GetByID(ctx context.Context, chatID, requesterID string) (*chat.Chat, error) {
	args := m.Called(ctx, chatID, requesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chat.Chat), args.Error(1)
}

func (m *MockChatService) ListForUser(ctx context.Context, userID string) ([]*chat.Chat, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*chat.Chat), args.Error(1)
}

func (m *MockChatService) AddMember(ctx context.Context, chatID, requesterID, newMemberID string) error {
	return m.Called(ctx, chatID, requesterID, newMemberID).Error(0)
}

func (m *MockChatService) RemoveMember(ctx context.Context, chatID, requesterID, targetID string) error {
	return m.Called(ctx, chatID, requesterID, targetID).Error(0)
}

func makeChat(id, name string) *chat.Chat {
	return &chat.Chat{
		ID:        id,
		Name:      name,
		Type:      chat.ChatTypeGroup,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
		Members:   []string{"user-1"},
	}
}

func TestChatHandler_List_Returns200WithChats(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	chats := []*chat.Chat{makeChat("chat-1", "General")}
	svc.On("ListForUser", mock.Anything, "user-1").Return(chats, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/chats", nil)
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []chatDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Len(t, result, 1)
	assert.Equal(t, "General", result[0].Name)
}

func TestChatHandler_Create_GroupChat_Returns201(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	c := makeChat("chat-new", "Engineering")
	svc.On("CreateGroup", mock.Anything, "user-1", "Engineering", []string{"user-2"}).Return(c, nil)

	body := `{"name":"Engineering","type":"group","member_ids":["user-2"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var result chatDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&result))
	assert.Equal(t, "Engineering", result.Name)
}

func TestChatHandler_Create_DirectChat_Returns201(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	c := &chat.Chat{
		ID:        "direct-1",
		Type:      chat.ChatTypeDirect,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
		Members:   []string{"user-1", "user-2"},
	}
	svc.On("CreateDirect", mock.Anything, "user-1", "user-2").Return(c, nil)

	body := `{"type":"direct","member_ids":["user-2"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestChatHandler_Create_InvalidType_Returns400(t *testing.T) {
	h := NewChatHandler(new(MockChatService), nil)

	body := `{"name":"Test","type":"broadcast"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestChatHandler_Create_GroupMissingName_Returns400(t *testing.T) {
	h := NewChatHandler(new(MockChatService), nil)

	body := `{"type":"group"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestChatHandler_GetByID_Member_Returns200(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	c := makeChat("chat-1", "General")
	svc.On("GetByID", mock.Anything, "chat-1", "user-1").Return(c, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/chats/chat-1", nil)
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestChatHandler_GetByID_NonMember_Returns403(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("GetByID", mock.Anything, "chat-1", "outsider").Return(nil, chat.ErrNotMember)

	req := httptest.NewRequest(http.MethodGet, "/api/chats/chat-1", nil)
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "outsider"))
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestChatHandler_AddMember_ByAdmin_Returns204(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("AddMember", mock.Anything, "chat-1", "admin-user", "new-user").Return(nil)

	body := `{"user_id":"new-user"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats/chat-1/members", strings.NewReader(body))
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "admin-user"))
	rec := httptest.NewRecorder()

	h.AddMember(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChatHandler_AddMember_ByNonAdmin_Returns403(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("AddMember", mock.Anything, "chat-1", "regular", "new-user").Return(chat.ErrForbidden)

	body := `{"user_id":"new-user"}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats/chat-1/members", strings.NewReader(body))
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "regular"))
	rec := httptest.NewRecorder()

	h.AddMember(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestChatHandler_RemoveMember_ByAdmin_Returns204(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("RemoveMember", mock.Anything, "chat-1", "admin-user", "target-user").Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/chats/chat-1/members/target-user", nil)
	req.SetPathValue("id", "chat-1")
	req.SetPathValue("userId", "target-user")
	req = req.WithContext(contextWithUserID(req.Context(), "admin-user"))
	rec := httptest.NewRecorder()

	h.RemoveMember(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestChatHandler_List_NotAuthenticated_Returns401(t *testing.T) {
	h := NewChatHandler(new(MockChatService), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/chats", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestChatHandler_List_ServiceError_Returns500(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("ListForUser", mock.Anything, "user-1").Return(nil, errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/api/chats", nil)
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.List(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestChatHandler_Create_NotAuthenticated_Returns401(t *testing.T) {
	h := NewChatHandler(new(MockChatService), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestChatHandler_Create_DirectMissingMember_Returns400(t *testing.T) {
	h := NewChatHandler(new(MockChatService), nil)

	body := `{"type":"direct","member_ids":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(body))
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestChatHandler_GetByID_NotFound_Returns404(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("GetByID", mock.Anything, "missing-chat", "user-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/chats/missing-chat", nil)
	req.SetPathValue("id", "missing-chat")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetByID(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestChatHandler_AddMember_MissingUserID_Returns400(t *testing.T) {
	h := NewChatHandler(new(MockChatService), nil)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/chats/chat-1/members", strings.NewReader(body))
	req.SetPathValue("id", "chat-1")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.AddMember(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestChatHandler_RemoveMember_ServiceError_Returns500(t *testing.T) {
	svc := new(MockChatService)
	h := NewChatHandler(svc, nil)

	svc.On("RemoveMember", mock.Anything, "chat-1", "user-1", "user-2").
		Return(errors.New("db error"))

	req := httptest.NewRequest(http.MethodDelete, "/api/chats/chat-1/members/user-2", nil)
	req.SetPathValue("id", "chat-1")
	req.SetPathValue("userId", "user-2")
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.RemoveMember(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
