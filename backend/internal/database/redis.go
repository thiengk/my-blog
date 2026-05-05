package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/personal-blog/backend/internal/config"
)

// NewRedisClient creates a new Redis client configured for Upstash.
// It parses the Redis URL (supports TLS for Upstash format) and verifies
// connectivity with a ping before returning.
func NewRedisClient(ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	if cfg.RedisURL == "" {
		return nil, fmt.Errorf("redis URL is not configured")
	}

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Configure pool settings suitable for Upstash
	opts.PoolSize = 10
	opts.MinIdleConns = 2
	opts.MaxRetries = 3
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second

	client := redis.NewClient(opts)

	// Verify connectivity
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
