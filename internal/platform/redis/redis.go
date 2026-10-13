// Package redis provides the Redis client used for token storage and other
// ephemeral state.
package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/ryanpramasyana/gdc/internal/platform/config"
)

// New opens a Redis client.
func New(cfg config.Redis) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

// Ping verifies connectivity to Redis.
func Ping(ctx context.Context, client *redis.Client) error {
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: ping: %w", err)
	}
	return nil
}
