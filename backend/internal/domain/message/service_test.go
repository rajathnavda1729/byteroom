package message

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, msg *Message) error {
	return m.Called(ctx, msg).Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Message), args.Error(1)
}

func (m *MockRepository) FindByChatID(ctx context.Context, chatID string, limit, offset int) ([]*Message, error) {
	args := m.Called(ctx, chatID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Message), args.Error(1)
}

func (m *MockRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type MockMembershipChecker struct {
	mock.Mock
}

func (m *MockMembershipChecker) IsMember(ctx context.Context, chatID, userID string) (bool, error) {
	args := m.Called(ctx, chatID, userID)
	return args.Bool(0), args.Error(1)
}

type MockSanitizer struct {
	mock.Mock
}

func (m *MockSanitizer) Sanitize(content string) string {
	return m.Called(content).String(0)
}

// --- Send tests ---

func TestMessageService_Send_ValidMessage_ReturnsMessageID(t *testing.T) {
	repo := new(MockRepository)
	membership := new(MockMembershipChecker)
	sanitizer := new(MockSanitizer)
	svc := NewService(repo, membership, sanitizer)

	membership.On("IsMember", mock.Anything, "chat-1", "user-1").Return(true, nil)
	sanitizer.On("Sanitize", "**Hello, world!**").Return("**Hello, world!**")
	repo.On("ExistsByID", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*message.Message")).Return(nil)

	msg := &Message{
		ChatID:      "chat-1",
		SenderID:    "user-1",
		ContentType: ContentTypeMarkdown,
		Content:     "**Hello, world!**",
	}
	result, err := svc.Send(context.Background(), msg)

	require.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.False(t, result.CreatedAt.IsZero())
	repo.AssertExpectations(t)
}

func TestMessageService_Send_NonMember_ReturnsForbidden(t *testing.T) {
	repo := new(MockRepository)
	membership := new(MockMembershipChecker)
	svc := NewService(repo, membership, nil)

	membership.On("IsMember", mock.Anything, "chat-1", "outsider").Return(false, nil)

	msg := &Message{ChatID: "chat-1", SenderID: "outsider", Content: "Hi"}
	_, err := svc.Send(context.Background(), msg)

	assert.ErrorIs(t, err, ErrForbidden)
	repo.AssertNotCalled(t, "Save")
}

func TestMessageService_Send_SanitizesContent(t *testing.T) {
	repo := new(MockRepository)
	membership := new(MockMembershipChecker)
	sanitizer := new(MockSanitizer)
	svc := NewService(repo, membership, sanitizer)

	membership.On("IsMember", mock.Anything, "chat-1", "user-1").Return(true, nil)
	sanitizer.On("Sanitize", `<script>alert("xss")</script>Hello`).Return("Hello")
	repo.On("ExistsByID", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*message.Message")).Return(nil)

	msg := &Message{
		ChatID:   "chat-1",
		SenderID: "user-1",
		Content:  `<script>alert("xss")</script>Hello`,
	}
	result, err := svc.Send(context.Background(), msg)

	require.NoError(t, err)
	assert.Equal(t, "Hello", result.Content)
	sanitizer.AssertExpectations(t)
}

func TestMessageService_Send_EmptyContent_ReturnsError(t *testing.T) {
	repo := new(MockRepository)
	membership := new(MockMembershipChecker)
	sanitizer := new(MockSanitizer)
	svc := NewService(repo, membership, sanitizer)

	membership.On("IsMember", mock.Anything, "chat-1", "user-1").Return(true, nil)
	sanitizer.On("Sanitize", "").Return("")

	msg := &Message{ChatID: "chat-1", SenderID: "user-1", Content: ""}
	_, err := svc.Send(context.Background(), msg)

	assert.ErrorIs(t, err, ErrInvalidContent)
}

func TestMessageService_Send_DuplicateID_IsIdempotent(t *testing.T) {
	repo := new(MockRepository)
	membership := new(MockMembershipChecker)
	sanitizer := new(MockSanitizer)
	svc := NewService(repo, membership, sanitizer)

	existing := &Message{ID: "msg-1", ChatID: "chat-1", SenderID: "user-1", Content: "Hi"}
	membership.On("IsMember", mock.Anything, "chat-1", "user-1").Return(true, nil)
	sanitizer.On("Sanitize", "Hi").Return("Hi")
	repo.On("ExistsByID", mock.Anything, "msg-1").Return(true, nil)
	repo.On("FindByID", mock.Anything, "msg-1").Return(existing, nil)

	msg := &Message{ID: "msg-1", ChatID: "chat-1", SenderID: "user-1", Content: "Hi"}
	result, err := svc.Send(context.Background(), msg)

	require.NoError(t, err)
	assert.Equal(t, "msg-1", result.ID)
	repo.AssertNotCalled(t, "Save")
}

func TestMessageService_GetHistory_DefaultsLimit_WhenZero(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil, nil)

	repo.On("FindByChatID", mock.Anything, "chat-1", 50, 0).Return([]*Message{}, nil)

	_, err := svc.GetHistory(context.Background(), "chat-1", 0, 0)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestMessageService_GetHistory_CapsLimit_AboveMax(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo, nil, nil)

	repo.On("FindByChatID", mock.Anything, "chat-1", 50, 0).Return([]*Message{}, nil)

	_, err := svc.GetHistory(context.Background(), "chat-1", 200, 0)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
