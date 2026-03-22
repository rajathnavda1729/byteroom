package api

import (
	"net/http"

	"github.com/byteroom/backend/internal/api/handler"
	"github.com/byteroom/backend/internal/api/middleware"
)

// Router wires all HTTP routes and applies the middleware chain.
type Router struct {
	auth    *handler.AuthHandler
	chats   *handler.ChatHandler
	msgs    *handler.MessageHandler
	uploads *handler.UploadHandler
	health  *handler.HealthHandler
	ws      *handler.WSHandler
	jwt     middleware.TokenValidator
	origins []string
}

// RouterConfig holds the handler dependencies needed to build the router.
type RouterConfig struct {
	Auth    *handler.AuthHandler
	Chats   *handler.ChatHandler
	Msgs    *handler.MessageHandler
	Uploads *handler.UploadHandler
	Health  *handler.HealthHandler
	WS      *handler.WSHandler
	JWT     middleware.TokenValidator
	Origins []string
}

// NewRouter creates a Router from the provided config.
func NewRouter(cfg RouterConfig) *Router {
	return &Router{
		auth:    cfg.Auth,
		chats:   cfg.Chats,
		msgs:    cfg.Msgs,
		uploads: cfg.Uploads,
		health:  cfg.Health,
		ws:      cfg.WS,
		jwt:     cfg.JWT,
		origins: cfg.Origins,
	}
}

// Build assembles and returns the http.Handler for the entire API.
func (r *Router) Build() http.Handler {
	mux := http.NewServeMux()

	// Public
	mux.HandleFunc("GET /health", r.health.Check)
	mux.HandleFunc("POST /api/auth/register", r.auth.Register)
	mux.HandleFunc("POST /api/auth/login", r.auth.Login)

	// Authenticated
	authMW := middleware.Auth(r.jwt)

	mux.Handle("GET /api/users/me", authMW(http.HandlerFunc(r.auth.Me)))
	mux.Handle("GET /api/users/search", authMW(http.HandlerFunc(r.auth.SearchUsers)))

	mux.Handle("GET /api/chats", authMW(http.HandlerFunc(r.chats.List)))
	mux.Handle("POST /api/chats", authMW(http.HandlerFunc(r.chats.Create)))
	mux.Handle("GET /api/chats/{id}", authMW(http.HandlerFunc(r.chats.GetByID)))
	mux.Handle("POST /api/chats/{id}/members", authMW(http.HandlerFunc(r.chats.AddMember)))
	mux.Handle("DELETE /api/chats/{id}/members/{userId}", authMW(http.HandlerFunc(r.chats.RemoveMember)))

	mux.Handle("GET /api/chats/{id}/messages", authMW(http.HandlerFunc(r.msgs.GetHistory)))
	mux.Handle("POST /api/chats/{id}/messages", authMW(http.HandlerFunc(r.msgs.Send)))

	mux.Handle("POST /api/upload/request", authMW(http.HandlerFunc(r.uploads.RequestUploadURL)))

	// WebSocket — authenticates via query-param token (not the Auth middleware)
	if r.ws != nil {
		mux.Handle("GET /ws", r.ws)
	}

	// Apply global middleware chain: RequestID → Security headers → CORS → Logger
	return middleware.RequestID(
		middleware.SecurityHeaders()(
			middleware.CORS(r.origins)(
				middleware.Logger(mux),
			),
		),
	)
}
