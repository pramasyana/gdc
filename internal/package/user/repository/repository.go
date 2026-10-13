// Package repository implements the user module's repository ports on top of
// MySQL via GORM. It joins the active transaction transparently by pulling it
// out of the context.
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
	"github.com/ryanpramasyana/gdc/internal/package/user/ports"
	"github.com/ryanpramasyana/gdc/internal/platform/database"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

type userRepository struct {
	db *gorm.DB
}

// New builds a MySQL-backed ports.UserRepository.
func New(db *gorm.DB) ports.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	m := userToModel(user)
	if err := database.FromContext(ctx, r.db).Create(m).Error; err != nil {
		if database.IsDuplicate(err) {
			return kernel.ErrDuplicate
		}
		return err
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*domain.User, error) {
	var m userModel
	if err := database.FromContext(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		return nil, database.Translate(err)
	}
	return userToDomain(&m), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	if err := database.FromContext(ctx, r.db).First(&m, "email = ?", email).Error; err != nil {
		return nil, database.Translate(err)
	}
	return userToDomain(&m), nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	err := database.FromContext(ctx, r.db).Model(&userModel{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"name":  user.Name,
			"email": user.Email,
		}).Error
	return database.Translate(err)
}
