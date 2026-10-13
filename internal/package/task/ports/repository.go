package ports

import (
	"context"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
)

// TaskRepository persists tasks.
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id uint64) (*domain.Task, error)
	List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int64, error)
	Update(ctx context.Context, task *domain.Task) error
	UpdateAssignee(ctx context.Context, taskID, assigneeID uint64) error
	Delete(ctx context.Context, id uint64) error
}

// TaskLogRepository persists task audit logs.
type TaskLogRepository interface {
	Create(ctx context.Context, log *domain.TaskLog) error
}
