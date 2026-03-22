package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeJWT(t *testing.T, userID string) string {
	t.Helper()
	mgr := NewJWTManager("test-secret", 24*time.Hour)
	token, err := mgr.Generate(userID)
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}
	return token
}

func TestAuthMiddleware_ValidToken_SetsUserIDInContext(t *testing.T) {
	mgr := NewJWTManager("test-secret", 24*time.Hour)
	token := makeJWT(t, "user-123")

	var capturedID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID, _ = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	Auth(mgr)(inner).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-123", capturedID)
}

func TestAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	mgr := NewJWTManager("test-secret", 24*time.Hour)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	Auth(mgr)(inner).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	mgr := NewJWTManager("test-secret", 24*time.Hour)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()

	Auth(mgr)(inner).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuthMiddleware_ExpiredToken_Returns401(t *testing.T) {
	expiredMgr := NewJWTManager("test-secret", -1*time.Hour)
	token, _ := expiredMgr.Generate("user-123")

	validMgr := NewJWTManager("test-secret", 24*time.Hour)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	Auth(validMgr)(inner).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUserIDFromContext_WithValue_ReturnsID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, "user-abc", id)
	})

	mgr := NewJWTManager("test-secret", 24*time.Hour)
	token, _ := mgr.Generate("user-abc")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	Auth(mgr)(inner).ServeHTTP(rec, req)
}
