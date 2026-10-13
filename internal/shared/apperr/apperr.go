// Package apperr defines the application's typed errors. Every business or
// transport failure is expressed as an *Error carrying an HTTP status and a
// stable machine-readable code so clients receive a consistent envelope.
package apperr

import (
	"errors"
	"net/http"
)

// Stable error codes returned to clients.
const (
	CodeValidation            = "VALIDATION_ERROR"
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeForbidden             = "FORBIDDEN"
	CodeNotFound              = "NOT_FOUND"
	CodeConflict              = "CONFLICT"
	CodeUnprocessable         = "UNPROCESSABLE_ENTITY"
	CodeInternal              = "INTERNAL_ERROR"
	CodeRateLimited           = "RATE_LIMITED"
	CodeIdempotencyRequired   = "IDEMPOTENCY_KEY_REQUIRED"
	CodeIdempotencyInvalid    = "IDEMPOTENCY_KEY_INVALID"
	CodeIdempotencyMismatch   = "IDEMPOTENCY_KEY_MISMATCH"
	CodeIdempotencyInProgress = "IDEMPOTENCY_KEY_IN_PROGRESS"
)

// Error is the canonical application error.
type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Err     error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

// Unwrap exposes the wrapped cause for errors.Is/errors.As.
func (e *Error) Unwrap() error { return e.Err }

// WithCause returns a copy of the error wrapping cause.
func (e *Error) WithCause(cause error) *Error {
	clone := *e
	clone.Err = cause
	return &clone
}

// WithFields returns a copy of the error carrying per-field details, typically
// validation messages keyed by request field name.
func (e *Error) WithFields(fields map[string]string) *Error {
	clone := *e
	clone.Fields = fields
	return &clone
}

// New builds an Error with the given status, code and message.
func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

// BadRequest returns a 400 error.
func BadRequest(code, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}

// Unauthorized returns a 401 error.
func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden returns a 403 error.
func Forbidden(message string) *Error {
	return New(http.StatusForbidden, CodeForbidden, message)
}

// NotFound returns a 404 error.
func NotFound(message string) *Error {
	return New(http.StatusNotFound, CodeNotFound, message)
}

// Conflict returns a 409 error.
func Conflict(code, message string) *Error {
	return New(http.StatusConflict, code, message)
}

// Unprocessable returns a 422 error.
func Unprocessable(message string) *Error {
	return New(http.StatusUnprocessableEntity, CodeUnprocessable, message)
}

// Internal returns a 500 error that never exposes internal details.
func Internal(cause error) *Error {
	return New(http.StatusInternalServerError, CodeInternal, "internal server error").WithCause(cause)
}

// From converts any error into an *Error. Unknown errors become 500 responses
// with a generic message so that internals never leak to clients.
func From(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
