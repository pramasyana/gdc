// Package response defines the task module's outbound HTTP payloads.
package response

import (
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/task/domain"
)

// TaskResponse is a serialized task.
type TaskResponse struct {
	ID          uint64     `json:"id" example:"1"`
	OwnerID     uint64     `json:"owner_id" example:"1"`
	AssigneeID  *uint64    `json:"assignee_id" example:"2"`
	TeamID      uint64     `json:"team_id" example:"1"`
	Title       string     `json:"title" example:"Write the report"`
	Description string     `json:"description" example:"Quarterly report"`
	Status      string     `json:"status" example:"todo"`
	DueDate     *time.Time `json:"due_date" example:"2026-01-10T17:00:00Z"`
	CreatedAt   time.Time  `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt   time.Time  `json:"updated_at" example:"2026-01-02T15:04:05Z"`
}

// FromTask maps a domain task to its response form.
func FromTask(task *domain.Task) TaskResponse {
	return TaskResponse{
		ID:          task.ID,
		OwnerID:     task.OwnerID,
		AssigneeID:  task.AssigneeID,
		TeamID:      task.TeamID,
		Title:       task.Title,
		Description: task.Description,
		Status:      string(task.Status),
		DueDate:     task.DueDate,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

// FromTasks maps a slice of domain tasks to response forms.
func FromTasks(tasks []*domain.Task) []TaskResponse {
	responses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, FromTask(task))
	}
	return responses
}
