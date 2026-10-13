package apperr

import (
	"errors"

	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

// TranslateNotFound converts the repository not-found sentinel into the
// resource's application error, leaving any other error untouched.
func TranslateNotFound(err error, resource string) error {
	if errors.Is(err, kernel.ErrNotFound) {
		return NotFound(resource + " not found")
	}
	return err
}

// TranslateConflict converts the repository duplicate sentinel into a conflict,
// leaving any other error untouched.
func TranslateConflict(err error, message string) error {
	if errors.Is(err, kernel.ErrDuplicate) {
		return Conflict(CodeConflict, message)
	}
	return err
}
