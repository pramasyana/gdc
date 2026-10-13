package ports

import (
	"context"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
)

// UpdateProfileInput is the data required to update a profile.
type UpdateProfileInput struct {
	Name  string
	Email string
}

// UserUsecase defines user profile operations.
type UserUsecase interface {
	GetProfile(ctx context.Context, userID uint64) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uint64, input UpdateProfileInput) (*domain.User, error)
}
