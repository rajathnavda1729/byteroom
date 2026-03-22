package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DBPinger is the minimal DB interface required by the health handler.
type DBPinger interface {
	PingContext(ctx context.Context) error
}

// HubStats is the minimal interface the health handler requires from the WS Hub.
type HubStats interface {
	ActiveConnections() int
}

// HealthHandler handles liveness and readiness endpoints.
type HealthHandler struct {
	db        DBPinger
	hub       HubStats
	version   string
	startTime time.Time
}

// NewHealthHandler creates a HealthHandler.
// db and hub may be nil; if so those checks are skipped.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{startTime: time.Now(), version: "1.0.0"}
}

// WithDB attaches a database connection for the DB health check.
func (h *HealthHandler) WithDB(db DBPinger) *HealthHandler {
	h.db = db
	return h
}

// WithHub attaches the WebSocket hub for connection-count reporting.
func (h *HealthHandler) WithHub(hub HubStats) *HealthHandler {
	h.hub = hub
	return h
}

type healthResponse struct {
	Status  string            `json:"status"`
	Version string            `json:"version"`
	Uptime  string            `json:"uptime"`
	Checks  map[string]string `json:"checks"`
}

// Check handles GET /health.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)

	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.db.PingContext(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
		} else {
			checks["database"] = "healthy"
		}
	}

	if h.hub != nil {
		checks["websocket_hub"] = fmt.Sprintf("healthy: %d connections", h.hub.ActiveConnections())
	}

	status := "healthy"
	code := http.StatusOK
	for _, v := range checks {
		if !strings.HasPrefix(v, "healthy") {
			status = "unhealthy"
			code = http.StatusServiceUnavailable
			break
		}
	}

	resp := healthResponse{
		Status:  status,
		Version: h.version,
		Uptime:  time.Since(h.startTime).String(),
		Checks:  checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(resp)
}
