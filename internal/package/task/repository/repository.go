// Package repository implements the task module's repository ports on top of
// MySQL via GORM.
package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
	"github.com/ryanpramasyana/gdc/internal/package/task/ports"
	"github.com/ryanpramasyana/gdc/internal/platform/database"
	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

type taskRepository struct {
	db *gorm.DB
}

// New builds a MySQL-backed ports.TaskRepository.
func New(db *gorm.DB) ports.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
	m := &taskModel{
		OwnerID:     task.OwnerID,
		AssigneeID:  task.AssigneeID,
		TeamID:      task.TeamID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		DueDate:     task.DueDate,
	}
	if err := database.FromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	task.ID = m.ID
	task.CreatedAt = m.CreatedAt
	task.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uint64) (*domain.Task, error) {
	var m taskModel
	if err := database.FromContext(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		return nil, database.Translate(err)
	}
	return taskToDomain(&m), nil
}

func (r *taskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int64, error) {
	query := database.FromContext(ctx, r.db).Model(&taskModel{})

	if filter.OwnerID != nil {
		query = query.Where("owner_id = ? OR assignee_id = ?", *filter.OwnerID, *filter.OwnerID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", string(filter.Status))
	}
	if filter.Search != "" {
		query = query.Where("title LIKE ?", "%"+filter.Search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []taskModel
	if err := query.
		Order("created_at DESC").
		Offset(filter.Offset()).
		Limit(filter.Limit).
		Find(&models).Error; err != nil {
		return nil, 0, err
	}

	tasks := make([]*domain.Task, 0, len(models))
	for i := range models {
		tasks = append(tasks, taskToDomain(&models[i]))
	}
	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *domain.Task) error {
	result := database.FromContext(ctx, r.db).Model(&taskModel{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"title":       task.Title,
			"description": task.Description,
			"status":      string(task.Status),
			"due_date":    task.DueDate,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return kernel.ErrNotFound
	}
	return nil
}

func (r *taskRepository) UpdateAssignee(ctx context.Context, taskID, assigneeID uint64) error {
	result := database.FromContext(ctx, r.db).Model(&taskModel{}).
		Where("id = ?", taskID).
		Update("assignee_id", assigneeID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return kernel.ErrNotFound
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id uint64) error {
	result := database.FromContext(ctx, r.db).Delete(&taskModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return kernel.ErrNotFound
	}
	return nil
}

type taskLogRepository struct {
	db *gorm.DB
}

// NewTaskLog builds a MySQL-backed ports.TaskLogRepository.
func NewTaskLog(db *gorm.DB) ports.TaskLogRepository {
	return &taskLogRepository{db: db}
}

func (r *taskLogRepository) Create(ctx context.Context, log *domain.TaskLog) error {
	m := taskLogToModel(log)
	if err := database.FromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	log.ID = m.ID
	log.CreatedAt = m.CreatedAt
	return nil
}
