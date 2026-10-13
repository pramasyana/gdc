// Package middleware contains the HTTP middleware chain: request identity,
// structured logging, panic recovery, authentication and CORS.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/shared/reqctx"
)

// Context keys set by middleware and read by handlers.
const (
	ContextAccessJTI = "access_jti"
	ContextAccessExp = "access_exp"
)

// UserID returns the authenticated user identifier from the request context.
func UserID(c *gin.Context) (uint64, bool) {
	return reqctx.UserID(c.Request.Context())
}

// bearer extracts a bearer token from the Authorization header.
func bearer(c *gin.Context) string {
	const prefix = "Bearer "
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
