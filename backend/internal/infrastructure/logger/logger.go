package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

// Logger wraps slog with context-aware helpers.
type Logger struct {
	sl *slog.Logger
}

// New creates a Logger. level is "debug"|"info"|"warn"|"error".
// format is "json" (default) or "text".
func New(level, format string) *Logger {
	var l slog.Level
	_ = l.UnmarshalText([]byte(level))

	opts := &slog.HandlerOptions{Level: l}
	var h slog.Handler
	if format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}
	return &Logger{sl: slog.New(h)}
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, args ...any) { l.sl.Info(msg, args...) }

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, args ...any) { l.sl.Warn(msg, args...) }

// Error logs at ERROR level, adding "error" to the attrs.
func (l *Logger) Error(msg string, err error, args ...any) {
	args = append(args, "error", err)
	l.sl.Error(msg, args...)
}

// WithContext returns a child Logger pre-filled with request_id / user_id from ctx.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	sl := l.sl
	if rid, ok := ctx.Value(RequestIDKey).(string); ok && rid != "" {
		sl = sl.With("request_id", rid)
	}
	if uid, ok := ctx.Value(UserIDKey).(string); ok && uid != "" {
		sl = sl.With("user_id", uid)
	}
	return &Logger{sl: sl}
}

// With returns a child Logger with fixed key/value pairs.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{sl: l.sl.With(args...)}
}
