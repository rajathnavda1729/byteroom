package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/byteroom/backend/internal/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type MockUserService struct{ mock.Mock }

func (m *MockUserService) Register(ctx context.Context, username, displayName, password string) (*user.User, error) {
	args := m.Called(ctx, username, displayName, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Authenticate(ctx context.Context, username, password string) (*user.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserService) Search(ctx context.Context, query, excludeUserID string, limit int) ([]*user.User, error) {
	args := m.Called(ctx, query, excludeUserID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

type MockTokenIssuer struct{ mock.Mock }

func (m *MockTokenIssuer) Generate(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// --- Register tests ---

func TestAuthHandler_Register_ValidRequest_Returns201WithToken(t *testing.T) {
	svc := new(MockUserService)
	tokens := new(MockTokenIssuer)
	h := NewAuthHandler(svc, tokens)

	u := &user.User{ID: "user-1", Username: "alice", DisplayName: "Alice"}
	svc.On("Register", mock.Anything, "alice", "Alice", "s3cr3t!").Return(u, nil)
	tokens.On("Generate", "user-1").Return("jwt.token.here", nil)

	body := `{"username":"alice","display_name":"Alice","password":"s3cr3t!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var resp authResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "jwt.token.here", resp.Token)
	assert.Equal(t, "alice", resp.User.Username)
}

func TestAuthHandler_Register_DuplicateUsername_Returns409(t *testing.T) {
	svc := new(MockUserService)
	tokens := new(MockTokenIssuer)
	h := NewAuthHandler(svc, tokens)

	svc.On("Register", mock.Anything, "alice", "", "pass").
		Return(nil, user.ErrDuplicateUsername)

	body := `{"username":"alice","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAuthHandler_Register_MissingFields_Returns400(t *testing.T) {
	h := NewAuthHandler(new(MockUserService), new(MockTokenIssuer))

	body := `{"username":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Register_InvalidJSON_Returns400(t *testing.T) {
	h := NewAuthHandler(new(MockUserService), new(MockTokenIssuer))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader("bad json"))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Login tests ---

func TestAuthHandler_Login_ValidCredentials_Returns200WithToken(t *testing.T) {
	svc := new(MockUserService)
	tokens := new(MockTokenIssuer)
	h := NewAuthHandler(svc, tokens)

	u := &user.User{ID: "user-1", Username: "alice"}
	svc.On("Authenticate", mock.Anything, "alice", "s3cr3t!").Return(u, nil)
	tokens.On("Generate", "user-1").Return("jwt.token.here", nil)

	body := `{"username":"alice","password":"s3cr3t!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp authResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	assert.Equal(t, "jwt.token.here", resp.Token)
}

func TestAuthHandler_Login_InvalidCredentials_Returns401(t *testing.T) {
	svc := new(MockUserService)
	tokens := new(MockTokenIssuer)
	h := NewAuthHandler(svc, tokens)

	svc.On("Authenticate", mock.Anything, "alice", "wrong").
		Return(nil, user.ErrInvalidCredentials)

	body := `{"username":"alice","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_Login_MissingFields_Returns400(t *testing.T) {
	h := NewAuthHandler(new(MockUserService), new(MockTokenIssuer))

	body := `{"username":"alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- Me tests ---

func TestAuthHandler_Me_Authenticated_ReturnsUser(t *testing.T) {
	svc := new(MockUserService)
	h := NewAuthHandler(svc, new(MockTokenIssuer))

	u := &user.User{ID: "user-1", Username: "alice", DisplayName: "Alice"}
	svc.On("GetByID", mock.Anything, "user-1").Return(u, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var dto userDTO
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&dto))
	assert.Equal(t, "alice", dto.Username)
}

func TestAuthHandler_Me_UserNotFound_Returns404(t *testing.T) {
	svc := new(MockUserService)
	h := NewAuthHandler(svc, new(MockTokenIssuer))

	svc.On("GetByID", mock.Anything, "user-1").Return(nil, user.ErrUserNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	req = req.WithContext(contextWithUserID(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAuthHandler_Me_NotAuthenticated_Returns401(t *testing.T) {
	h := NewAuthHandler(new(MockUserService), new(MockTokenIssuer))

	req := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --- edge cases ---

func TestAuthHandler_Login_ServiceError_Returns401(t *testing.T) {
	svc := new(MockUserService)
	tokens := new(MockTokenIssuer)
	h := NewAuthHandler(svc, tokens)

	svc.On("Authenticate", mock.Anything, "alice", "pass").
		Return(nil, errors.New("db error"))

	body := `{"username":"alice","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthHandler_Login_InvalidJSON_Returns400(t *testing.T) {
	h := NewAuthHandler(new(MockUserService), new(MockTokenIssuer))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("bad json"))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAuthHandler_Register_InvalidUsername_Returns400(t *testing.T) {
	svc := new(MockUserService)
	h := NewAuthHandler(svc, new(MockTokenIssuer))

	svc.On("Register", mock.Anything, "ab", "", "pass").
		Return(nil, user.ErrInvalidUsername)

	body := `{"username":"ab","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// --- service error coverage ---

func TestAuthHandler_Register_ServiceError_Returns500(t *testing.T) {
	svc := new(MockUserService)
	h := NewAuthHandler(svc, new(MockTokenIssuer))

	svc.On("Register", mock.Anything, "alice", "", "pass").
		Return(nil, errors.New("db down"))

	body := `{"username":"alice","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
