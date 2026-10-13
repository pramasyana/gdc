// Package response defines the user module's outbound HTTP payloads.
package response

import (
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
)

// UserResponse is a serialized user.
type UserResponse struct {
	ID        uint64    `json:"id" example:"1"`
	Name      string    `json:"name" example:"Jane Doe"`
	Email     string    `json:"email" example:"jane@example.com"`
	CreatedAt time.Time `json:"created_at" example:"2026-01-02T15:04:05Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-01-02T15:04:05Z"`
}

// FromUser maps a domain user to its response form.
func FromUser(user *domain.User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
