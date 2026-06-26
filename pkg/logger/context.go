package logger

import "context"

type contextKey string

const traceIDKey contextKey = "trace_id"

// WithTraceID returns a new context carrying the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID extracts the trace ID from context. Returns "no-trace" if absent.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok && id != "" {
		return id
	}
	return "no-trace"
}
