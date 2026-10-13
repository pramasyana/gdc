package ports

import (
	"context"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
)

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uint64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}
