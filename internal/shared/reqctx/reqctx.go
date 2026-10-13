// Package reqctx carries per-request values across layers using the standard
// context, keeping transport concerns out of the domain and usecase packages.
package reqctx

import "context"

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// WithRequestID returns a context carrying the given request identifier.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request identifier stored in the context, if any.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithUserID returns a context carrying the authenticated user identifier.
func WithUserID(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserID returns the authenticated user identifier stored in the context.
func UserID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(userIDKey).(uint64)
	return id, ok
}
