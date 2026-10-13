// Package httpx holds small HTTP helpers shared by every feature module's
// handler: request binding and request-context accessors.
package httpx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
)

// Bind decodes and validates a JSON request body, returning a friendly
// application error on failure. Validation failures carry per-field messages.
func Bind(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil {
		return validationError(target, err)
	}
	return nil
}

// BindQuery decodes and validates query parameters.
func BindQuery(c *gin.Context, target any) error {
	if err := c.ShouldBindQuery(target); err != nil {
		return validationError(target, err)
	}
	return nil
}

func validationError(target any, err error) *apperr.Error {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return apperr.BadRequest(apperr.CodeValidation, err.Error())
	}

	fields := make(map[string]string, len(validationErrs))
	messages := make([]string, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		msg := fieldMessage(fieldErr)
		fields[jsonFieldName(target, fieldErr.Field())] = msg
		messages = append(messages, msg)
	}
	return apperr.BadRequest(apperr.CodeValidation, strings.Join(messages, "; ")).WithFields(fields)
}

// jsonFieldName maps a struct field name to its JSON key so validation errors
// are keyed the same way the client sent them.
func jsonFieldName(target any, goField string) string {
	t := reflect.TypeOf(target)
	for t != nil && t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return strings.ToLower(goField)
	}
	field, ok := t.FieldByName(goField)
	if !ok {
		return strings.ToLower(goField)
	}
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" || name == "-" {
		return strings.ToLower(goField)
	}
	return name
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "oneof":
		return fe.Field() + " must be one of: " + fe.Param()
	default:
		return fe.Field() + " is invalid"
	}
}
