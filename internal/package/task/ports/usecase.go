package ports

import (
	"context"
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
)

// CreateTaskInput is the data required to create a task.
type CreateTaskInput struct {
	Title       string
	Description string
	Status      domain.TaskStatus
	DueDate     *time.Time
}

// UpdateTaskInput is the data required to update a task.
type UpdateTaskInput struct {
	Title       string
	Description string
	Status      domain.TaskStatus
	DueDate     *time.Time
}

// ListTasksInput constrains a task listing.
type ListTasksInput struct {
	Status domain.TaskStatus
	Search string
	Page   int
	Limit  int
}

// RenderFunc serializes a task into the exact HTTP payload that will be stored
// for idempotent replays. The transport layer supplies it so the usecase stays
// free of serialization concerns.
type RenderFunc func(task *domain.Task) (status int, body []byte)

// IdempotentResult is the outcome of an idempotent operation. When Replayed is
// true the payload was produced by an earlier request and no new side effects
// occurred.
type IdempotentResult struct {
	Status   int
	Body     []byte
	Replayed bool
}

// TaskUsecase defines task management operations.
type TaskUsecase interface {
	Create(ctx context.Context, userID uint64, idempotencyKey string, input CreateTaskInput, render RenderFunc) (*IdempotentResult, error)
	List(ctx context.Context, userID uint64, input ListTasksInput) ([]*domain.Task, int64, error)
	Get(ctx context.Context, userID, taskID uint64) (*domain.Task, error)
	Update(ctx context.Context, userID, taskID uint64, input UpdateTaskInput) (*domain.Task, error)
	Delete(ctx context.Context, userID, taskID uint64) error
	Assign(ctx context.Context, actorID, taskID, assigneeID uint64) (*domain.Task, error)
}
