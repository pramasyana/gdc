package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ryanpramasyana/gdc/internal/shared/reqctx"
)

// Logger emits one structured JSON log line per request. Level is INFO for
// 2xx/3xx, WARN for 4xx and ERROR for 5xx.
func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		level := slog.LevelInfo
		switch {
		case status >= 500:
			level = slog.LevelError
		case status >= 400:
			level = slog.LevelWarn
		}

		attrs := []any{
			"request_id", reqctx.RequestID(c.Request.Context()),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}
		if userID, ok := UserID(c); ok {
			attrs = append(attrs, "user_id", userID)
		}
		logger.Log(c.Request.Context(), level, "request completed", attrs...)
	}
}
