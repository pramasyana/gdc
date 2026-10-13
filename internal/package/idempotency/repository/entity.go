package repository

import (
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/idempotency/domain"
)

// idempotencyModel is the GORM representation of an idempotency record.
type idempotencyModel struct {
	ID             uint64    `gorm:"column:id;primaryKey"`
	Key            string    `gorm:"column:idempotency_key"`
	UserID         uint64    `gorm:"column:user_id"`
	RequestMethod  string    `gorm:"column:request_method"`
	RequestPath    string    `gorm:"column:request_path"`
	RequestHash    string    `gorm:"column:request_hash"`
	State          string    `gorm:"column:state"`
	ResponseStatus int       `gorm:"column:response_status"`
	ResponseBody   []byte    `gorm:"column:response_body"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
}

func (idempotencyModel) TableName() string { return "idempotency_keys" }

func idempotencyToModel(r *domain.IdempotencyRecord) *idempotencyModel {
	return &idempotencyModel{
		Key:           r.Key,
		UserID:        r.UserID,
		RequestMethod: r.Method,
		RequestPath:   r.Path,
		RequestHash:   r.RequestHash,
		State:         string(r.State),
		ExpiresAt:     r.ExpiresAt,
	}
}

func idempotencyToDomain(m *idempotencyModel) *domain.IdempotencyRecord {
	return &domain.IdempotencyRecord{
		Key:            m.Key,
		UserID:         m.UserID,
		Method:         m.RequestMethod,
		Path:           m.RequestPath,
		RequestHash:    m.RequestHash,
		State:          domain.IdempotencyState(m.State),
		ResponseStatus: m.ResponseStatus,
		ResponseBody:   m.ResponseBody,
		CreatedAt:      m.CreatedAt,
		ExpiresAt:      m.ExpiresAt,
	}
}
