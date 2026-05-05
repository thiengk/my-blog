package middleware

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitConfig holds configuration for a specific endpoint group.
type RateLimitConfig struct {
	WindowSize  time.Duration // Sliding window size
	MaxRequests int64         // Maximum requests allowed within the window
	FailOpen    bool          // Allow requests when Redis is unavailable
}

// RateLimiter implements sliding window rate limiting using Redis sorted sets.
type RateLimiter struct {
	client  *redis.Client
	configs map[string]RateLimitConfig
}

// NewRateLimiter creates a new RateLimiter with the given Redis client and
// per-endpoint-group configurations.
func NewRateLimiter(client *redis.Client, configs map[string]RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		client:  client,
		configs: configs,
	}
}

// Allow checks whether a request from the given IP to the endpoint group
// is permitted under the rate limit. It returns:
//   - allowed: whether the request should proceed
//   - retryAfter: how long the client should wait before retrying (only meaningful when denied)
//   - err: any error encountered (Redis errors, etc.)
//
// Sliding window algorithm using Redis sorted sets:
//  1. Key: rate:{ip}:{endpoint_group}
//  2. Score: timestamp in Unix milliseconds
//  3. Member: unique request ID (timestamp + random suffix)
//  4. On each request:
//     - Remove entries older than window start (ZREMRANGEBYSCORE)
//     - Count remaining entries (ZCARD)
//     - If count < max_requests: add new entry (ZADD), allow
//     - If count >= max_requests: deny, return Retry-After
//  5. Set TTL on key = window_size * 2
func (rl *RateLimiter) Allow(ip string, endpointGroup string) (bool, time.Duration, error) {
	cfg, ok := rl.configs[endpointGroup]
	if !ok {
		// No config for this endpoint group, allow by default
		return true, 0, nil
	}

	// If Redis client is nil, apply fail-open policy
	if rl.client == nil {
		if cfg.FailOpen {
			log.Printf("WARNING: Rate limiter fail-open: Redis client is nil for %s:%s", ip, endpointGroup)
			return true, 0, nil
		}
		return false, cfg.WindowSize, fmt.Errorf("redis client is nil and fail-open is disabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	now := time.Now()
	nowMs := now.UnixMilli()
	windowStart := now.Add(-cfg.WindowSize).UnixMilli()
	key := fmt.Sprintf("rate:%s:%s", ip, endpointGroup)

	// Generate unique member: timestamp + random suffix to avoid collisions
	member := fmt.Sprintf("%d:%d", nowMs, rand.Int63())

	// Use a pipeline for atomic operations
	pipe := rl.client.Pipeline()

	// Remove entries older than the window start
	pipe.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))

	// Count current entries in the window
	countCmd := pipe.ZCard(ctx, key)

	// Execute pipeline to get current count
	_, err := pipe.Exec(ctx)
	if err != nil {
		return rl.handleRedisError(err, cfg, ip, endpointGroup)
	}

	currentCount := countCmd.Val()

	// Check if the request exceeds the limit
	if currentCount >= cfg.MaxRequests {
		// Get the oldest entry to calculate retry-after
		oldestEntries, err := rl.client.ZRangeWithScores(ctx, key, 0, 0).Result()
		if err != nil || len(oldestEntries) == 0 {
			// Fallback: suggest waiting for the full window
			return false, cfg.WindowSize, nil
		}

		oldestTimestamp := int64(oldestEntries[0].Score)
		// The oldest entry will expire at oldestTimestamp + windowSize
		expiresAt := oldestTimestamp + cfg.WindowSize.Milliseconds()
		retryAfter := time.Duration(expiresAt-nowMs) * time.Millisecond

		if retryAfter <= 0 {
			retryAfter = time.Second
		}

		return false, retryAfter, nil
	}

	// Add the new entry and set TTL
	pipe2 := rl.client.Pipeline()
	pipe2.ZAdd(ctx, key, redis.Z{
		Score:  float64(nowMs),
		Member: member,
	})
	pipe2.Expire(ctx, key, cfg.WindowSize*2)

	_, err = pipe2.Exec(ctx)
	if err != nil {
		return rl.handleRedisError(err, cfg, ip, endpointGroup)
	}

	return true, 0, nil
}

// GetCount returns the current number of requests in the sliding window
// for the given IP and endpoint group.
func (rl *RateLimiter) GetCount(ip string, endpointGroup string) (int64, error) {
	cfg, ok := rl.configs[endpointGroup]
	if !ok {
		return 0, nil
	}

	if rl.client == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	now := time.Now()
	windowStart := now.Add(-cfg.WindowSize).UnixMilli()
	key := fmt.Sprintf("rate:%s:%s", ip, endpointGroup)

	// Remove expired entries first
	rl.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", windowStart))

	// Count remaining entries
	count, err := rl.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get rate limit count: %w", err)
	}

	return count, nil
}

// Middleware returns a Gin middleware handler that applies rate limiting
// for the specified endpoint group.
func (rl *RateLimiter) Middleware(endpointGroup string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := extractClientIP(c)

		allowed, retryAfter, err := rl.Allow(ip, endpointGroup)
		if err != nil {
			// Error occurred but fail-open allowed the request through
			// (if fail-open denied, allowed would be false)
			log.Printf("WARNING: Rate limiter error for %s on %s: %v", ip, endpointGroup, err)
		}

		if !allowed {
			retryAfterSeconds := int(retryAfter.Seconds())
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}

			c.Header("Retry-After", fmt.Sprintf("%d", retryAfterSeconds))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": retryAfterSeconds,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// handleRedisError applies the fail-open policy when a Redis error occurs.
func (rl *RateLimiter) handleRedisError(err error, cfg RateLimitConfig, ip string, endpointGroup string) (bool, time.Duration, error) {
	if cfg.FailOpen {
		log.Printf("WARNING: Rate limiter fail-open due to Redis error for %s:%s: %v", ip, endpointGroup, err)
		return true, 0, err
	}
	return false, cfg.WindowSize, fmt.Errorf("redis error: %w", err)
}

// extractClientIP extracts the real client IP from the request,
// checking X-Forwarded-For, X-Real-IP headers, and falling back to RemoteAddr.
func extractClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (may contain multiple IPs: client, proxy1, proxy2)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP (the original client)
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (strip port if present)
	remoteAddr := c.Request.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		return remoteAddr
	}
	return host
}
