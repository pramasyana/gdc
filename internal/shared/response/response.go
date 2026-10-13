// Package response renders the API's uniform JSON envelope for both success
// and error responses.
package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
)

// ── Response types ─────────────────────────────────────

// SuccessResponse is the envelope for a successful single-resource response.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ErrorResponse is the envelope for every failure. Errors carries per-field
// validation details when available.
type ErrorResponse struct {
	Status    int               `json:"status"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Errors    map[string]string `json:"errors,omitempty"`
}

// ListResponse is the envelope for a paginated collection.
type ListResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	Data     any      `json:"data"`
	Metadata Metadata `json:"metadata"`
}

// Metadata describes the pagination of a ListResponse.
type Metadata struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

// ── Constructors ───────────────────────────────────────

// NewSuccess builds a success envelope.
func NewSuccess(msg string, data any) SuccessResponse {
	return SuccessResponse{Success: true, Message: msg, Data: data}
}

// NewError builds an error envelope without field details.
func NewError(msg string) ErrorResponse {
	return ErrorResponse{Status: http.StatusInternalServerError, Code: apperr.CodeInternal, Message: msg, Timestamp: time.Now().UTC().Format(time.RFC3339)}
}

// NewErrorWithErrors builds an error envelope carrying per-field details.
func NewErrorWithErrors(msg string, fields map[string]string) ErrorResponse {
	return ErrorResponse{Status: http.StatusBadRequest, Code: apperr.CodeValidation, Message: msg, Timestamp: time.Now().UTC().Format(time.RFC3339), Errors: fields}
}

// NewList builds a paginated collection envelope.
func NewList(msg string, data any, page, perPage int, total int64) ListResponse {
	if perPage < 1 {
		perPage = 1
	}
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return ListResponse{
		Success: true,
		Message: msg,
		Data:    data,
		Metadata: Metadata{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       total,
			TotalPages:  totalPages,
		},
	}
}

// ── gin shortcuts ──────────────────────────────────────

// OK writes a 200 success response.
func OK(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, NewSuccess(msg, data))
}

// Created writes a 201 success response.
func Created(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusCreated, NewSuccess(msg, data))
}

// List writes a 200 success response with pagination metadata.
func List(c *gin.Context, msg string, data any, page, perPage int, total int64) {
	c.JSON(http.StatusOK, NewList(msg, data, page, perPage, total))
}

// BadRequest writes a 400 error response.
func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, NewError(msg))
}

// BadRequestWithErrors writes a 400 error response with per-field details.
func BadRequestWithErrors(c *gin.Context, msg string, fields map[string]string) {
	c.JSON(http.StatusBadRequest, NewErrorWithErrors(msg, fields))
}

// Unauthorized writes a 401 error response.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, NewError(msg))
}

// Forbidden writes a 403 error response.
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, NewError(msg))
}

// NotFound writes a 404 error response.
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, NewError(msg))
}

// Conflict writes a 409 error response.
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, NewError(msg))
}

// TooManyRequests writes a 429 error response.
func TooManyRequests(c *gin.Context, msg string) {
	c.JSON(http.StatusTooManyRequests, NewError(msg))
}

// ServerError writes a 500 error response.
func ServerError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, NewError(msg))
}

// Error renders any application error using the standard error envelope. When
// the error carries per-field validation details they are included.
func Error(c *gin.Context, err error) {
	appErr := apperr.From(err)
	if len(appErr.Fields) > 0 {
		r := NewErrorWithErrors(appErr.Message, appErr.Fields)
		r.Status = appErr.Status
		r.Code = appErr.Code
		r.Timestamp = time.Now().UTC().Format(time.RFC3339)
		c.AbortWithStatusJSON(appErr.Status, r)
		return
	}
	r := NewError(appErr.Message)
	r.Status = appErr.Status
	r.Code = appErr.Code
	r.Timestamp = time.Now().UTC().Format(time.RFC3339)
	c.AbortWithStatusJSON(appErr.Status, r)
}

// Raw writes a pre-rendered response body verbatim. It is used to replay
// idempotent responses exactly as they were first produced.
func Raw(c *gin.Context, status int, body []byte) {
	c.Data(status, "application/json; charset=utf-8", body)
}
