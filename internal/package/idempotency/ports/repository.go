package ports

import (
	"context"

	"github.com/ryanpramasyana/gdc/internal/package/idempotency/domain"
)

// IdempotencyRepository persists idempotency records. Acquire must be atomic:
// it attempts to insert a processing record and, on a uniqueness conflict,
// returns the existing record with acquired=false.
type IdempotencyRepository interface {
	Acquire(ctx context.Context, record *domain.IdempotencyRecord) (existing *domain.IdempotencyRecord, acquired bool, err error)
	GetByKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error)
	Complete(ctx context.Context, key string, status int, body []byte) error
}
