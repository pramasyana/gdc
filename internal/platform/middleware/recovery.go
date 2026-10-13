package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
	"github.com/ryanpramasyana/gdc/internal/shared/reqctx"
	"github.com/ryanpramasyana/gdc/internal/shared/response"
)

// Recovery catches panics anywhere in the chain, logs the stack server-side
// and returns a generic 500 response without leaking internals.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.ErrorContext(c.Request.Context(), "panic recovered",
					"request_id", reqctx.RequestID(c.Request.Context()),
					"panic", fmt.Sprintf("%v", recovered),
					"stack", string(debug.Stack()),
				)
				response.Error(c, apperr.Internal(fmt.Errorf("panic: %v", recovered)))
			}
		}()
		c.Next()
	}
}
