package kernel

import "context"

// Notification is a message to be delivered to a user. Delivery may be mocked.
type Notification struct {
	UserID   uint64
	Type     string
	Title    string
	Body     string
	Metadata map[string]any
}

// Notifier delivers notifications. Implementations may log instead of sending.
type Notifier interface {
	Send(ctx context.Context, notification Notification) error
}
