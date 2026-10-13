package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
)

// userModel is the GORM representation of a user.
type userModel struct {
	ID            uint64         `gorm:"column:id;primaryKey"`
	Name          string         `gorm:"column:name"`
	Email         string         `gorm:"column:email"`
	PasswordHash  string         `gorm:"column:password_hash"`
	PrimaryTeamID *uint64        `gorm:"column:primary_team_id"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (userModel) TableName() string { return "users" }

func userToModel(u *domain.User) *userModel {
	return &userModel{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		PasswordHash:  u.PasswordHash,
		PrimaryTeamID: u.PrimaryTeamID,
	}
}

func userToDomain(m *userModel) *domain.User {
	return &domain.User{
		ID:            m.ID,
		Name:          m.Name,
		Email:         m.Email,
		PasswordHash:  m.PasswordHash,
		PrimaryTeamID: m.PrimaryTeamID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
