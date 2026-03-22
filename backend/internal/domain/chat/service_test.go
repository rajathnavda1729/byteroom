package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Save(ctx context.Context, c *Chat) error {
	return m.Called(ctx, c).Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*Chat, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Chat), args.Error(1)
}

func (m *MockRepository) GetUserChatIDs(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) FindByMember(ctx context.Context, userID string) ([]*Chat, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Chat), args.Error(1)
}

func (m *MockRepository) AddMember(ctx context.Context, chatID, userID string, role MemberRole) error {
	return m.Called(ctx, chatID, userID, role).Error(0)
}

func (m *MockRepository) RemoveMember(ctx context.Context, chatID, userID string) error {
	return m.Called(ctx, chatID, userID).Error(0)
}

func (m *MockRepository) GetMemberRole(ctx context.Context, chatID, userID string) (MemberRole, error) {
	args := m.Called(ctx, chatID, userID)
	return args.Get(0).(MemberRole), args.Error(1)
}

func (m *MockRepository) IsMember(ctx context.Context, chatID, userID string) (bool, error) {
	args := m.Called(ctx, chatID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRepository) FindDirectBetween(ctx context.Context, userID1, userID2 string) (*Chat, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Chat), args.Error(1)
}

// --- CreateGroup ---

func TestChatService_CreateGroup_ValidInput_AddsCreatorAsAdmin(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("Save", mock.Anything, mock.AnythingOfType("*chat.Chat")).Return(nil)
	repo.On("AddMember", mock.Anything, mock.AnythingOfType("string"), "user-1", RoleAdmin).Return(nil)
	repo.On("AddMember", mock.Anything, mock.AnythingOfType("string"), "user-2", RoleMember).Return(nil)

	c, err := svc.CreateGroup(context.Background(), "user-1", "Engineering", []string{"user-2"})

	require.NoError(t, err)
	assert.Equal(t, "Engineering", c.Name)
	assert.Equal(t, ChatTypeGroup, c.Type)
	repo.AssertExpectations(t)
}

func TestChatService_CreateGroup_EmptyName_ReturnsError(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	_, err := svc.CreateGroup(context.Background(), "user-1", "", nil)

	assert.ErrorIs(t, err, ErrEmptyChatName)
	repo.AssertNotCalled(t, "Save")
}

// --- AddMember ---

func TestChatService_AddMember_ByAdmin_Succeeds(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetMemberRole", mock.Anything, "chat-1", "admin-user").Return(RoleAdmin, nil)
	repo.On("AddMember", mock.Anything, "chat-1", "new-user", RoleMember).Return(nil)

	err := svc.AddMember(context.Background(), "chat-1", "admin-user", "new-user")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestChatService_AddMember_ByNonAdmin_ReturnsForbidden(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetMemberRole", mock.Anything, "chat-1", "regular-user").Return(RoleMember, nil)

	err := svc.AddMember(context.Background(), "chat-1", "regular-user", "new-user")

	assert.ErrorIs(t, err, ErrForbidden)
	repo.AssertNotCalled(t, "AddMember")
}

func TestChatService_AddMember_NonMember_ReturnsNotMember(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetMemberRole", mock.Anything, "chat-1", "outsider").
		Return(MemberRole(""), errors.New("not found"))

	err := svc.AddMember(context.Background(), "chat-1", "outsider", "someone")

	assert.ErrorIs(t, err, ErrNotMember)
}

// --- RemoveMember ---

func TestChatService_RemoveMember_AdminRemovesOther_Succeeds(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetMemberRole", mock.Anything, "chat-1", "admin-user").Return(RoleAdmin, nil)
	repo.On("RemoveMember", mock.Anything, "chat-1", "target-user").Return(nil)

	err := svc.RemoveMember(context.Background(), "chat-1", "admin-user", "target-user")

	assert.NoError(t, err)
}

func TestChatService_RemoveMember_MemberLeavesOwn_Succeeds(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetMemberRole", mock.Anything, "chat-1", "user-1").Return(RoleMember, nil)
	repo.On("RemoveMember", mock.Anything, "chat-1", "user-1").Return(nil)

	err := svc.RemoveMember(context.Background(), "chat-1", "user-1", "user-1")

	assert.NoError(t, err)
}

func TestChatService_RemoveMember_NonAdminRemovesOther_ReturnsForbidden(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	repo.On("GetMemberRole", mock.Anything, "chat-1", "regular-user").Return(RoleMember, nil)

	err := svc.RemoveMember(context.Background(), "chat-1", "regular-user", "other-user")

	assert.ErrorIs(t, err, ErrForbidden)
}

// --- GetByID ---

func TestChatService_GetByID_MemberRequester_ReturnsChat(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	stored := &Chat{ID: "chat-1", Name: "Room", Type: ChatTypeGroup}
	repo.On("FindByID", mock.Anything, "chat-1").Return(stored, nil)
	repo.On("IsMember", mock.Anything, "chat-1", "user-1").Return(true, nil)

	c, err := svc.GetByID(context.Background(), "chat-1", "user-1")

	require.NoError(t, err)
	assert.Equal(t, "chat-1", c.ID)
}

func TestChatService_GetByID_NonMemberRequester_ReturnsNotMember(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	stored := &Chat{ID: "chat-1"}
	repo.On("FindByID", mock.Anything, "chat-1").Return(stored, nil)
	repo.On("IsMember", mock.Anything, "chat-1", "outsider").Return(false, nil)

	_, err := svc.GetByID(context.Background(), "chat-1", "outsider")

	assert.ErrorIs(t, err, ErrNotMember)
}

// --- CreateDirect ---

func TestChatService_CreateDirect_TwoUsers_Succeeds(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	// No existing direct chat
	repo.On("FindDirectBetween", mock.Anything, "u1", "u2").Return(nil, ErrChatNotFound)
	repo.On("Save", mock.Anything, mock.AnythingOfType("*chat.Chat")).Return(nil)
	repo.On("AddMember", mock.Anything, mock.AnythingOfType("string"), "u1", RoleMember).Return(nil)
	repo.On("AddMember", mock.Anything, mock.AnythingOfType("string"), "u2", RoleMember).Return(nil)

	c, err := svc.CreateDirect(context.Background(), "u1", "u2")

	require.NoError(t, err)
	assert.Equal(t, ChatTypeDirect, c.Type)
}

func TestChatService_CreateDirect_ExistingChat_ReturnsExisting(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	existing := &Chat{ID: "existing-room", Type: ChatTypeDirect}
	repo.On("FindDirectBetween", mock.Anything, "u1", "u2").Return(existing, nil)

	c, err := svc.CreateDirect(context.Background(), "u1", "u2")

	require.NoError(t, err)
	assert.Equal(t, "existing-room", c.ID)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

// --- ListForUser ---

func TestChatService_ListForUser_ReturnsChats(t *testing.T) {
	repo := new(MockRepository)
	svc := NewService(repo)

	chats := []*Chat{{ID: "c1"}, {ID: "c2"}}
	repo.On("FindByMember", mock.Anything, "user-1").Return(chats, nil)

	result, err := svc.ListForUser(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// --- Entity ---

func TestChatEntity_IsMember_KnownMember_ReturnsTrue(t *testing.T) {
	c := &Chat{Members: []string{"u1", "u2"}}
	assert.True(t, c.IsMember("u1"))
}

func TestChatEntity_IsMember_UnknownMember_ReturnsFalse(t *testing.T) {
	c := &Chat{Members: []string{"u1"}}
	assert.False(t, c.IsMember("u99"))
}

func TestChatEntity_Validate_DirectChat_WrongMemberCount_ReturnsError(t *testing.T) {
	c := &Chat{Type: ChatTypeDirect, Members: []string{"u1"}}
	assert.ErrorIs(t, c.Validate(), ErrDirectChatSize)
}

func TestChatEntity_Validate_GroupChat_ValidNameNoError(t *testing.T) {
	c := &Chat{Type: ChatTypeGroup, Name: "Test", Members: []string{"u1"}}
	assert.NoError(t, c.Validate())
}

func TestChatEntity_Validate_InvalidType_ReturnsError(t *testing.T) {
	c := &Chat{Type: "broadcast"}
	assert.ErrorIs(t, c.Validate(), ErrInvalidChatType)
}

func TestChatEntity_Validate_DirectChat_TwoMembers_NoError(t *testing.T) {
	c := &Chat{Type: ChatTypeDirect, Members: []string{"u1", "u2"}}
	assert.NoError(t, c.Validate())
}
