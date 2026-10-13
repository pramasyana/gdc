// Package request defines the auth module's inbound HTTP payloads.
package request

// RegisterRequest is the registration payload.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=120" example:"Jane Doe"`
	Email    string `json:"email" binding:"required,email" example:"jane@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"secret123"`
	TeamName string `json:"team_name" binding:"omitempty,max=120" example:"Platform"`
}

// LoginRequest is the login payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"jane@example.com"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// RefreshRequest is the token refresh payload.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOi..."`
}

// LogoutRequest optionally carries the refresh token to revoke.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOi..."`
}
