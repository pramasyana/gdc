// Package cache implements the domain token store on top of Redis.
package cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	refreshPrefix   = "auth:refresh:"
	blacklistPrefix = "auth:blacklist:"
)

// TokenStore stores refresh-token state and access-token blacklists in Redis.
type TokenStore struct {
	client *redis.Client
}

// NewTokenStore builds a TokenStore.
func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

// SaveRefresh records a refresh token identifier for the given user.
func (s *TokenStore) SaveRefresh(ctx context.Context, jti string, userID uint64, ttl time.Duration) error {
	if err := s.client.Set(ctx, refreshPrefix+jti, strconv.FormatUint(userID, 10), ttl).Err(); err != nil {
		return fmt.Errorf("cache: save refresh: %w", err)
	}
	return nil
}

// HasRefresh reports whether a refresh token identifier is still valid.
func (s *TokenStore) HasRefresh(ctx context.Context, jti string) (bool, error) {
	n, err := s.client.Exists(ctx, refreshPrefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("cache: has refresh: %w", err)
	}
	return n > 0, nil
}

// RevokeRefresh invalidates a refresh token identifier.
func (s *TokenStore) RevokeRefresh(ctx context.Context, jti string) error {
	if err := s.client.Del(ctx, refreshPrefix+jti).Err(); err != nil {
		return fmt.Errorf("cache: revoke refresh: %w", err)
	}
	return nil
}

// BlacklistAccess records an access token identifier as revoked until its ttl.
func (s *TokenStore) BlacklistAccess(ctx context.Context, jti string, ttl time.Duration) error {
	if err := s.client.Set(ctx, blacklistPrefix+jti, "1", ttl).Err(); err != nil {
		return fmt.Errorf("cache: blacklist access: %w", err)
	}
	return nil
}

// IsAccessBlacklisted reports whether an access token identifier was revoked.
func (s *TokenStore) IsAccessBlacklisted(ctx context.Context, jti string) (bool, error) {
	n, err := s.client.Exists(ctx, blacklistPrefix+jti).Result()
	if err != nil {
		return false, fmt.Errorf("cache: is access blacklisted: %w", err)
	}
	return n > 0, nil
}
