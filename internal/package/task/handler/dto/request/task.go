// Package request defines the task module's inbound HTTP payloads.
package request

import "time"

// CreateTaskRequest is the payload for creating a task.
type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255" example:"Write the report"`
	Description string     `json:"description" binding:"omitempty,max=5000" example:"Quarterly report"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress done cancelled" example:"todo"`
	DueDate     *time.Time `json:"due_date" example:"2026-01-10T17:00:00Z"`
}

// UpdateTaskRequest is the payload for updating a task.
type UpdateTaskRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255" example:"Write the report"`
	Description string     `json:"description" binding:"omitempty,max=5000" example:"Quarterly report"`
	Status      string     `json:"status" binding:"omitempty,oneof=todo in_progress done cancelled" example:"in_progress"`
	DueDate     *time.Time `json:"due_date" example:"2026-01-10T17:00:00Z"`
}

// AssignTaskRequest is the payload for assigning a task.
type AssignTaskRequest struct {
	AssigneeID uint64 `json:"assignee_id" binding:"required" example:"2"`
}

// ListTasksQuery holds the task listing filters.
type ListTasksQuery struct {
	Status string `form:"status" binding:"omitempty,oneof=todo in_progress done cancelled" example:"todo"`
	Search string `form:"search" example:"report"`
	Page   int    `form:"page" binding:"omitempty,min=1" example:"1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
}
