package ports

import (
	"context"
	"time"

	"github.com/ryanpramasyana/gdc/internal/package/user/domain"
)

// RegisterInput is the data required to create an account.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
	TeamName string
}

// LoginInput is the data required to authenticate.
type LoginInput struct {
	Email    string
	Password string
}

// AuthResult is the outcome of a successful authentication flow.
type AuthResult struct {
	User             *domain.User
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// AuthUsecase defines authentication operations.
type AuthUsecase interface {
	Register(ctx context.Context, input RegisterInput) (*AuthResult, error)
	Login(ctx context.Context, input LoginInput) (*AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, userID uint64, accessJTI string, accessExpiry time.Time, refreshToken string) error
}
