// Package response defines the auth module's outbound HTTP payloads.
package response

import (
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/auth/ports"
	userresp "github.com/ryanpramasyana/gdc/internal/package/user/handler/dto/response"
)

// AuthResponse is a serialized authentication result.
type AuthResponse struct {
	User             userresp.UserResponse `json:"user"`
	AccessToken      string                `json:"access_token" example:"eyJhbGciOi..."`
	RefreshToken     string                `json:"refresh_token" example:"eyJhbGciOi..."`
	AccessExpiresAt  time.Time             `json:"access_expires_at" example:"2026-01-02T15:19:05Z"`
	RefreshExpiresAt time.Time             `json:"refresh_expires_at" example:"2026-01-09T15:04:05Z"`
}

// FromAuthResult maps an authentication result to its response form.
func FromAuthResult(result *ports.AuthResult) AuthResponse {
	return AuthResponse{
		User:             userresp.FromUser(result.User),
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		AccessExpiresAt:  result.AccessExpiresAt,
		RefreshExpiresAt: result.RefreshExpiresAt,
	}
}
