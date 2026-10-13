package repository

import "time"

// teamModel is the GORM representation of a team.
type teamModel struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (teamModel) TableName() string { return "teams" }

// teamMemberModel is the GORM representation of a team membership.
type teamMemberModel struct {
	TeamID    uint64    `gorm:"column:team_id;primaryKey"`
	UserID    uint64    `gorm:"column:user_id;primaryKey"`
	Role      string    `gorm:"column:role"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (teamMemberModel) TableName() string { return "team_members" }
