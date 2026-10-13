// Package domain holds the idempotency module's entities. It imports no
// framework.
package domain

import "time"

// IdempotencyState tracks the lifecycle of an idempotency key.
type IdempotencyState string

// Supported idempotency states.
const (
	IdempotencyStateProcessing IdempotencyState = "processing"
	IdempotencyStateCompleted  IdempotencyState = "completed"
)

// IdempotencyRecord stores the outcome of an idempotent request so replays can
// return the exact same response without side effects.
type IdempotencyRecord struct {
	Key            string
	UserID         uint64
	Method         string
	Path           string
	RequestHash    string
	State          IdempotencyState
	ResponseStatus int
	ResponseBody   []byte
	CreatedAt      time.Time
	ExpiresAt      time.Time
}

// IsCompleted reports whether the request finished and a response is stored.
func (r *IdempotencyRecord) IsCompleted() bool {
	return r.State == IdempotencyStateCompleted
}

// IsExpired reports whether the record fell outside the replay window.
func (r *IdempotencyRecord) IsExpired(now time.Time) bool {
	return now.After(r.ExpiresAt)
}
