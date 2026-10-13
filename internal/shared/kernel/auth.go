package kernel

import (
	"context"
	"time"
)

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

// TokenType distinguishes access tokens from refresh tokens.
type TokenType string

// Supported token types.
const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// TokenClaims is the payload extracted from a signed token.
type TokenClaims struct {
	UserID uint64
	JTI    string
	Type   TokenType
	Expiry time.Time
}

// TokenProvider signs and parses JWT tokens.
type TokenProvider interface {
	Generate(userID uint64, tokenType TokenType, ttl time.Duration) (*TokenClaims, string, error)
	Parse(token string, tokenType TokenType) (*TokenClaims, error)
}

// TokenStore keeps refresh-token state and access-token blacklists in Redis.
type TokenStore interface {
	SaveRefresh(ctx context.Context, jti string, userID uint64, ttl time.Duration) error
	HasRefresh(ctx context.Context, jti string) (bool, error)
	RevokeRefresh(ctx context.Context, jti string) error
	BlacklistAccess(ctx context.Context, jti string, ttl time.Duration) error
	IsAccessBlacklisted(ctx context.Context, jti string) (bool, error)
}
