// Package repository implements the idempotency module's repository port on top
// of MySQL via GORM.
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/package/idempotency/domain"
	"github.com/ryanpramasyana/gdc/internal/package/idempotency/ports"
	"github.com/ryanpramasyana/gdc/internal/platform/database"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

type idempotencyRepository struct {
	db *gorm.DB
}

// New builds a MySQL-backed ports.IdempotencyRepository.
func New(db *gorm.DB) ports.IdempotencyRepository {
	return &idempotencyRepository{db: db}
}

// Acquire atomically inserts a processing record. A uniqueness conflict means
// another request already claimed the key; the existing record is returned
// with acquired=false. MySQL keeps the transaction usable after this error, so
// the caller may branch within the same transaction.
func (r *idempotencyRepository) Acquire(ctx context.Context, record *domain.IdempotencyRecord) (*domain.IdempotencyRecord, bool, error) {
	m := idempotencyToModel(record)
	if err := database.FromContext(ctx, r.db).Create(m).Error; err != nil {
		if !database.IsDuplicate(err) {
			return nil, false, err
		}
		existing, getErr := r.GetByKey(ctx, record.Key)
		if getErr != nil {
			return nil, false, getErr
		}
		return existing, false, nil
	}
	record.CreatedAt = m.CreatedAt
	return nil, true, nil
}

func (r *idempotencyRepository) GetByKey(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	var m idempotencyModel
	if err := database.FromContext(ctx, r.db).First(&m, "idempotency_key = ?", key).Error; err != nil {
		return nil, database.Translate(err)
	}
	return idempotencyToDomain(&m), nil
}

func (r *idempotencyRepository) Complete(ctx context.Context, key string, status int, body []byte) error {
	result := database.FromContext(ctx, r.db).Model(&idempotencyModel{}).
		Where("idempotency_key = ?", key).
		Updates(map[string]any{
			"state":           string(domain.IdempotencyStateCompleted),
			"response_status": status,
			"response_body":   body,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return kernel.ErrNotFound
	}
	return nil
}
