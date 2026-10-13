// Package repository implements the team module's repository port on top of
// MySQL via GORM.
package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ryanpramasyana/gdc/internal/package/team/domain"
	"github.com/ryanpramasyana/gdc/internal/package/team/ports"
	"github.com/ryanpramasyana/gdc/internal/platform/database"
)

type teamRepository struct {
	db *gorm.DB
}

// New builds a MySQL-backed ports.TeamRepository.
func New(db *gorm.DB) ports.TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(ctx context.Context, team *domain.Team) error {
	m := &teamModel{Name: team.Name}
	if err := database.FromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	team.ID = m.ID
	team.CreatedAt = m.CreatedAt
	team.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *teamRepository) AddMember(ctx context.Context, teamID, userID uint64, role domain.TeamRole) error {
	m := &teamMemberModel{TeamID: teamID, UserID: userID, Role: string(role)}
	return database.FromContext(ctx, r.db).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(m).Error
}

func (r *teamRepository) IsMember(ctx context.Context, teamID, userID uint64) (bool, error) {
	var count int64
	err := database.FromContext(ctx, r.db).Model(&teamMemberModel{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
