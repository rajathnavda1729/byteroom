package handler

import (
	"context"

	"github.com/byteroom/backend/internal/api/middleware"
)

// contextWithUserID injects a userID into a context for handler tests.
func contextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, middleware.UserIDKey, userID)
}
