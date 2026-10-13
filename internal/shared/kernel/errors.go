// Package kernel holds the cross-cutting contracts and primitives that every
// feature module and the platform adapters depend on. It imports no framework.
package kernel

import "errors"

// Sentinel errors returned by repositories. Usecases translate them into
// transport-agnostic application errors via the apperr helpers.
var (
	// ErrNotFound indicates the requested resource does not exist.
	ErrNotFound = errors.New("resource not found")
	// ErrDuplicate indicates a uniqueness constraint was violated.
	ErrDuplicate = errors.New("resource already exists")
)
