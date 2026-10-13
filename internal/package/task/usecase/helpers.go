package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
)

// validateIdempotencyKey ensures the header carries a well-formed UUID.
func validateIdempotencyKey(key string) error {
	if strings.TrimSpace(key) == "" {
		return apperr.BadRequest(apperr.CodeIdempotencyRequired, "Idempotency-Key header is required")
	}
	if _, err := uuid.Parse(key); err != nil {
		return apperr.BadRequest(apperr.CodeIdempotencyInvalid, "Idempotency-Key must be a valid UUID")
	}
	return nil
}

// hashCreateRequest derives a stable fingerprint of a create request so a key
// reused with a different payload can be rejected.
func hashCreateRequest(userID uint64, input ports.CreateTaskInput) string {
	payload := struct {
		UserID uint64
		Input  ports.CreateTaskInput
	}{UserID: userID, Input: input}
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

// normalizeStatus returns the provided status or the default, validating it.
func normalizeStatus(status domain.TaskStatus) (domain.TaskStatus, error) {
	if status == "" {
		return domain.TaskStatusTodo, nil
	}
	if !status.Valid() {
		return "", apperr.BadRequest(apperr.CodeValidation, "invalid task status")
	}
	return status, nil
}

// validateTitle ensures a task title is present and within bounds.
func validateTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return apperr.BadRequest(apperr.CodeValidation, "title is required")
	}
	if len([]rune(title)) > 255 {
		return apperr.BadRequest(apperr.CodeValidation, "title must be at most 255 characters")
	}
	return nil
}

// canAccess reports whether the user may read the task as owner or assignee.
func canAccess(userID uint64, task *domain.Task) bool {
	if task.OwnerID == userID {
		return true
	}
	return task.AssigneeID != nil && *task.AssigneeID == userID
}

// jsonString encodes a string as JSON so audit values are stored consistently.
func jsonString(value string) string {
	if value == "" {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}

// normalizePage clamps a page number to at least the default.
func normalizePage(page int) int {
	if page < 1 {
		return defaultPage
	}
	return page
}

// normalizeLimit clamps a page size into the allowed range.
func normalizeLimit(limit int) int {
	switch {
	case limit < 1:
		return defaultLimit
	case limit > maxLimit:
		return maxLimit
	default:
		return limit
	}
}
