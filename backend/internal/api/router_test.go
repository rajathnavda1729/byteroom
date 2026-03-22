package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/byteroom/backend/internal/api/handler"
	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/stretchr/testify/assert"
)

func buildTestRouter() http.Handler {
	jwt := middleware.NewJWTManager("test-secret", 24*time.Hour)
	return NewRouter(RouterConfig{
		Auth:    handler.NewAuthHandler(nil, nil),
		Chats:   handler.NewChatHandler(nil, nil),
		Msgs:    handler.NewMessageHandler(nil),
		Uploads: handler.NewUploadHandler(nil),
		Health:  handler.NewHealthHandler(),
		JWT:     jwt,
		Origins: []string{"http://localhost:5173"},
	}).Build()
}

func TestRouter_Health_Returns200(t *testing.T) {
	r := buildTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRouter_AuthEndpoints_ArePublic(t *testing.T) {
	r := buildTestRouter()

	for _, path := range []string{"/api/auth/register", "/api/auth/login"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		// Should not return 401 (public endpoints — 400 for bad body is fine)
		assert.NotEqual(t, http.StatusUnauthorized, rec.Code, "path=%s", path)
	}
}

func TestRouter_ProtectedEndpoints_Return401WithoutToken(t *testing.T) {
	r := buildTestRouter()

	protected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/users/me"},
		{http.MethodGet, "/api/chats"},
		{http.MethodPost, "/api/chats"},
		{http.MethodGet, "/api/chats/x/messages"},
		{http.MethodPost, "/api/chats/x/messages"},
		{http.MethodPost, "/api/upload/request"},
	}

	for _, tc := range protected {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s %s", tc.method, tc.path)
	}
}

func TestRouter_CORS_Headers_Present(t *testing.T) {
	r := buildTestRouter()

	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.NotEmpty(t, rec.Header().Get("Access-Control-Allow-Origin"))
}
