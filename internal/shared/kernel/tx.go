package kernel

import "context"

// TxManager runs a function inside a single database transaction. The
// transaction handle is propagated through the context so repositories
// participate automatically.
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
