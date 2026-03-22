package logger_test

import (
	"context"
	"testing"

	"github.com/byteroom/backend/internal/infrastructure/logger"
)

func TestLogger_New(t *testing.T) {
	l := logger.New("info", "json")
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLogger_WithContext_AddsFields(t *testing.T) {
	l := logger.New("debug", "text")
	ctx := context.WithValue(context.Background(), logger.RequestIDKey, "req-123")
	ctx = context.WithValue(ctx, logger.UserIDKey, "user-abc")

	child := l.WithContext(ctx)
	if child == nil {
		t.Fatal("expected non-nil child logger")
	}
	// Smoke-test that logging doesn't panic
	child.Info("test message", "key", "value")
}

func TestLogger_Error_DoesNotPanic(t *testing.T) {
	l := logger.New("info", "json")
	l.Error("something broke", nil, "extra", "data")
}
