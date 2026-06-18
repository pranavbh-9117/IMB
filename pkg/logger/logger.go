// Package logger provides a centralized structured logging mechanism for the application.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Init configures the global slog logger based on the environment.
// In production, it uses a strict JSON handler.
// In development, it uses a human-readable text handler.
func Init(env string) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "production" || env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

// Info logs an informational message. It extracts correlation IDs from the context if present.
func Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

// Error logs an error message. It extracts correlation IDs from the context if present.
func Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

// Warn logs a warning message. It extracts correlation IDs from the context if present.
func Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

// Debug logs a debug message. It extracts correlation IDs from the context if present.
func Debug(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}
