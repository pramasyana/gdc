// Package notify provides notification delivery implementations.
package notify

import (
	"context"
	"log/slog"

	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

// LogNotifier is a mock notifier that records notifications as structured
// logs instead of sending them externally.
type LogNotifier struct {
	logger *slog.Logger
}

// NewLogNotifier builds a LogNotifier.
func NewLogNotifier(logger *slog.Logger) *LogNotifier {
	return &LogNotifier{logger: logger}
}

// Send implements kernel.Notifier.
func (n *LogNotifier) Send(ctx context.Context, notification kernel.Notification) error {
	n.logger.InfoContext(ctx, "notification dispatched",
		"user_id", notification.UserID,
		"type", notification.Type,
		"title", notification.Title,
		"body", notification.Body,
	)
	return nil
}
