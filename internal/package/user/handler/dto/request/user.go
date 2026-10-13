// Package request defines the user module's inbound HTTP payloads.
package request

// UpdateProfileRequest is the payload for updating the current user's profile.
type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"required,min=1,max=120" example:"Jane Doe"`
	Email string `json:"email" binding:"required,email" example:"jane@example.com"`
}
