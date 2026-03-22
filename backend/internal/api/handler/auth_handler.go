package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/byteroom/backend/internal/api/middleware"
	"github.com/byteroom/backend/internal/domain/user"
)

// UserService is the interface the AuthHandler depends on.
type UserService interface {
	Register(ctx context.Context, username, displayName, password string) (*user.User, error)
	Authenticate(ctx context.Context, username, password string) (*user.User, error)
	GetByID(ctx context.Context, id string) (*user.User, error)
	Search(ctx context.Context, query, excludeID string, limit int) ([]*user.User, error)
}

// TokenIssuer generates JWT tokens.
type TokenIssuer interface {
	Generate(userID string) (string, error)
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	users  UserService
	tokens TokenIssuer
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(users UserService, tokens TokenIssuer) *AuthHandler {
	return &AuthHandler{users: users, tokens: tokens}
}

// registerRequest is the JSON body for POST /api/auth/register.
type registerRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// loginRequest is the JSON body for POST /api/auth/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      userDTO   `json:"user"`
}

type userDTO struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	u, err := h.users.Register(r.Context(), req.Username, req.DisplayName, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrDuplicateUsername):
			writeError(w, http.StatusConflict, "username already taken")
		case errors.Is(err, user.ErrInvalidUsername):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	token, err := h.tokens.Generate(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token generation failed")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User:      toUserDTO(u),
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	u, err := h.users.Authenticate(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.tokens.Generate(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token generation failed")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		User:      toUserDTO(u),
	})
}

// Me handles GET /api/users/me.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	u, err := h.users.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	writeJSON(w, http.StatusOK, toUserDTO(u))
}

// SearchUsers handles GET /api/users/search?q=<query>.
func (h *AuthHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	callerID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	q := r.URL.Query().Get("q")
	if len(q) < 2 {
		writeJSON(w, http.StatusOK, map[string]any{"users": []any{}})
		return
	}

	users, err := h.users.Search(r.Context(), q, callerID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed")
		return
	}

	dtos := make([]userDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toUserDTO(u))
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": dtos})
}

func toUserDTO(u *user.User) userDTO {
	return userDTO{
		UserID:      u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}
}
