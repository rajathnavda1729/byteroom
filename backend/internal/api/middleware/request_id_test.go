package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/byteroom/backend/internal/api/middleware"
)

func TestRequestID_GeneratesID(t *testing.T) {
	var capturedID string

	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = middleware.GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if capturedID == "" {
		t.Error("expected non-empty request ID in context")
	}
	if rec.Header().Get("X-Request-ID") != capturedID {
		t.Error("X-Request-ID response header should match context value")
	}
}

func TestRequestID_ReusesIncomingID(t *testing.T) {
	existing := "my-existing-id-123"

	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := middleware.GetRequestID(r.Context())
		if id != existing {
			t.Errorf("got %q, want %q", id, existing)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", existing)
	handler.ServeHTTP(httptest.NewRecorder(), req)
}
