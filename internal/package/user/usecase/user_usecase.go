// Package usecase implements the user module's business rules. It depends only
// on the module's domain and ports plus the shared kernel/errors.
package usecase

import (
	"context"
	"strings"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
	"github.com/ryanpramasyana/gdc/internal/package/user/ports"
	"github.com/ryanpramasyana/gdc/internal/shared/apperr"
)

type userUsecase struct {
	users ports.UserRepository
}

// New builds the user profile usecase.
func New(users ports.UserRepository) ports.UserUsecase {
	return &userUsecase{users: users}
}

// GetProfile returns the authenticated user's profile.
func (u *userUsecase) GetProfile(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.TranslateNotFound(err, "user")
	}
	return user, nil
}

// UpdateProfile updates the authenticated user's name and email.
func (u *userUsecase) UpdateProfile(ctx context.Context, userID uint64, input ports.UpdateProfileInput) (*domain.User, error) {
	name := strings.TrimSpace(input.Name)
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if name == "" || email == "" {
		return nil, apperr.BadRequest(apperr.CodeValidation, "name and email are required")
	}
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.TranslateNotFound(err, "user")
	}
	user.Name = name
	user.Email = email
	if err := u.users.Update(ctx, user); err != nil {
		return nil, apperr.TranslateConflict(err, "email is already in use")
	}
	return user, nil
}
