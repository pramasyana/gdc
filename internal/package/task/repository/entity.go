package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
)

// taskModel is the GORM representation of a task.
type taskModel struct {
	ID          uint64         `gorm:"column:id;primaryKey"`
	OwnerID     uint64         `gorm:"column:owner_id"`
	AssigneeID  *uint64        `gorm:"column:assignee_id"`
	TeamID      uint64         `gorm:"column:team_id"`
	Title       string         `gorm:"column:title"`
	Description string         `gorm:"column:description"`
	Status      string         `gorm:"column:status"`
	DueDate     *time.Time     `gorm:"column:due_date"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (taskModel) TableName() string { return "tasks" }

func taskToDomain(m *taskModel) *domain.Task {
	return &domain.Task{
		ID:          m.ID,
		OwnerID:     m.OwnerID,
		AssigneeID:  m.AssigneeID,
		TeamID:      m.TeamID,
		Title:       m.Title,
		Description: m.Description,
		Status:      domain.TaskStatus(m.Status),
		DueDate:     m.DueDate,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// taskLogModel is the GORM representation of a task audit entry.
type taskLogModel struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	TaskID    uint64    `gorm:"column:task_id"`
	ActorID   uint64    `gorm:"column:actor_id"`
	Action    string    `gorm:"column:action"`
	OldValue  *string   `gorm:"column:old_value"`
	NewValue  *string   `gorm:"column:new_value"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (taskLogModel) TableName() string { return "task_logs" }

func taskLogToModel(l *domain.TaskLog) *taskLogModel {
	m := &taskLogModel{TaskID: l.TaskID, ActorID: l.ActorID, Action: l.Action}
	if l.OldValue != "" {
		m.OldValue = &l.OldValue
	}
	if l.NewValue != "" {
		m.NewValue = &l.NewValue
	}
	return m
}
