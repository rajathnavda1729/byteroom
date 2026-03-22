package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/byteroom/backend/internal/api/handler"
)

func TestHealthHandler_Check(t *testing.T) {
	h := handler.NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body["status"] != "healthy" {
		t.Errorf("status = %q, want healthy", body["status"])
	}
	if body["version"] == "" {
		t.Error("version missing")
	}
	if body["uptime"] == "" {
		t.Error("uptime missing")
	}
}

type mockHub struct{ count int }

func (m *mockHub) ActiveConnections() int { return m.count }

func TestHealthHandler_WithHub(t *testing.T) {
	h := handler.NewHealthHandler().WithHub(&mockHub{count: 42})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Check(rec, req)

	var body map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&body)

	checks, ok := body["checks"].(map[string]interface{})
	if !ok {
		t.Fatal("checks field missing or wrong type")
	}
	if v, _ := checks["websocket_hub"].(string); v == "" {
		t.Error("websocket_hub check missing")
	}
}
