// Package logger builds the application's structured JSON logger.
package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON slog logger whose level depends on the environment:
// debug in development, info otherwise.
func New(env string) *slog.Logger {
	level := slog.LevelInfo
	if env != "production" {
		level = slog.LevelDebug
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
