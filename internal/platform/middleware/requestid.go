package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ryanpramasyana/gdc/internal/shared/reqctx"
)

// requestIDHeader is the header used to propagate the request identifier.
const requestIDHeader = "X-Request-ID"

// RequestID assigns a unique identifier to every request, honoring an inbound
// X-Request-ID when present, and exposes it on the response and context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Writer.Header().Set(requestIDHeader, id)
		c.Request = c.Request.WithContext(reqctx.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}
