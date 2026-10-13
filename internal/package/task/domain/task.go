// Package domain holds the task module's entities. It imports no framework.
package domain

import (
	"time"

	"github.com/ryanpramasyana/gdc/internal/shared/kernel"
)

// TaskStatus is the lifecycle state of a task.
type TaskStatus string

// Supported task statuses.
const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Valid reports whether the status is a known value.
func (s TaskStatus) Valid() bool {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// Task is a unit of work owned by a user and optionally assigned to another
// member of the same team.
type Task struct {
	ID          uint64
	OwnerID     uint64
	AssigneeID  *uint64
	TeamID      uint64
	Title       string
	Description string
	Status      TaskStatus
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TaskLog records an audit entry for a task mutation.
type TaskLog struct {
	ID        uint64
	TaskID    uint64
	ActorID   uint64
	Action    string
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}

// Task log actions.
const (
	ActionCreated  = "created"
	ActionUpdated  = "updated"
	ActionDeleted  = "deleted"
	ActionAssigned = "assigned"
)

// TaskFilter constrains a task listing query.
type TaskFilter struct {
	OwnerID *uint64
	Status  TaskStatus
	Search  string
	kernel.Pagination
}
