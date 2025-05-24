package ctxlog

import (
	"context"
	"log/slog"
)

type loggerKey struct{}

// WithLogger returns a new context with the provided logger attached.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// LoggerFromContext retrieves the logger stored in the context. If no logger
// is found, slog.Default() is returned.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
