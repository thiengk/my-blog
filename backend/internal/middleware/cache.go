package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// cachedResponse represents a cached HTTP response stored in Redis.
type cachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
}

// CacheMiddleware provides HTTP response caching using Redis.
type CacheMiddleware struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewCacheMiddleware creates a new CacheMiddleware with the given Redis client
// and default TTL for cached responses.
func NewCacheMiddleware(client *redis.Client, defaultTTL time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

// Middleware returns a Gin middleware handler that caches GET responses in Redis.
// The ttl parameter overrides the default TTL for this specific route.
// If ttl is 0, the default TTL is used.
//
// Behavior:
//   - Only caches GET requests
//   - Only caches successful responses (2xx status codes)
//   - Supports cache bypass via Cache-Control: no-cache header
//   - Fail-open: if Redis is unavailable, requests proceed without caching
func (cm *CacheMiddleware) Middleware(ttl time.Duration) gin.HandlerFunc {
	if ttl == 0 {
		ttl = cm.defaultTTL
	}

	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Support cache bypass via Cache-Control: no-cache header
		if c.GetHeader("Cache-Control") == "no-cache" {
			c.Next()
			return
		}

		// If Redis client is nil, skip caching (fail-open)
		if cm.client == nil {
			c.Next()
			return
		}

		cacheKey := cm.buildCacheKey(c)

		// Try to get cached response from Redis
		cached, err := cm.getFromCache(cacheKey)
		if err == nil && cached != nil {
			// Cache hit: return cached response directly
			for key, value := range cached.Headers {
				c.Header(key, value)
			}
			c.Header("X-Cache", "HIT")
			c.Data(cached.StatusCode, cached.Headers["Content-Type"], cached.Body)
			c.Abort()
			return
		}

		if err != nil && err != redis.Nil {
			// Redis error: log warning and proceed without caching (fail-open)
			log.Printf("WARNING: Cache middleware Redis read error: %v", err)
		}

		// Cache miss: execute handler and capture response
		c.Header("X-Cache", "MISS")
		writer := &responseCapture{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		// Only cache successful responses (2xx status codes)
		if writer.Status() >= 200 && writer.Status() < 300 {
			cm.storeInCache(cacheKey, writer, ttl)
		}
	}
}

// InvalidateCache deletes cache keys matching the given patterns.
// Patterns support Redis SCAN pattern matching (e.g., "cache:views:*").
// This is used for write-through cache invalidation when data changes.
func (cm *CacheMiddleware) InvalidateCache(patterns ...string) {
	if cm.client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, pattern := range patterns {
		// Use SCAN to find matching keys (safe for production, unlike KEYS)
		var cursor uint64
		for {
			keys, nextCursor, err := cm.client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				log.Printf("WARNING: Cache invalidation scan error for pattern %q: %v", pattern, err)
				break
			}

			if len(keys) > 0 {
				if err := cm.client.Del(ctx, keys...).Err(); err != nil {
					log.Printf("WARNING: Cache invalidation delete error: %v", err)
				}
			}

			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}
}

// InvalidateCacheKeys deletes specific cache keys directly (no pattern matching).
// Use this when you know the exact keys to invalidate.
func (cm *CacheMiddleware) InvalidateCacheKeys(keys ...string) {
	if cm.client == nil || len(keys) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := cm.client.Del(ctx, keys...).Err(); err != nil {
		log.Printf("WARNING: Cache key invalidation error: %v", err)
	}
}

// buildCacheKey generates a deterministic cache key from the request path
// and sorted query parameters.
// Format: "cache:" + path + "?" + sorted query params
func (cm *CacheMiddleware) buildCacheKey(c *gin.Context) string {
	path := c.Request.URL.Path

	// Sort query parameters for deterministic key generation
	queryParams := c.Request.URL.Query()
	if len(queryParams) == 0 {
		return "cache:" + path
	}

	// Sort parameter keys
	keys := make([]string, 0, len(queryParams))
	for k := range queryParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build sorted query string
	var sortedQuery strings.Builder
	for i, k := range keys {
		if i > 0 {
			sortedQuery.WriteString("&")
		}
		// Sort values for each key as well
		values := queryParams[k]
		sort.Strings(values)
		for j, v := range values {
			if j > 0 {
				sortedQuery.WriteString("&")
			}
			sortedQuery.WriteString(k)
			sortedQuery.WriteString("=")
			sortedQuery.WriteString(v)
		}
	}

	queryStr := sortedQuery.String()

	// If query string is too long, hash it to keep key size manageable
	if len(queryStr) > 128 {
		hash := sha256.Sum256([]byte(queryStr))
		return fmt.Sprintf("cache:%s?%x", path, hash[:8])
	}

	return "cache:" + path + "?" + queryStr
}

// getFromCache retrieves a cached response from Redis.
func (cm *CacheMiddleware) getFromCache(key string) (*cachedResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := cm.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var cached cachedResponse
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &cached, nil
}

// storeInCache stores a response in Redis with the given TTL.
func (cm *CacheMiddleware) storeInCache(key string, writer *responseCapture, ttl time.Duration) {
	cached := cachedResponse{
		StatusCode: writer.Status(),
		Headers:    make(map[string]string),
		Body:       writer.body.Bytes(),
	}

	// Capture relevant response headers
	for _, headerKey := range []string{"Content-Type", "Content-Length"} {
		if val := writer.Header().Get(headerKey); val != "" {
			cached.Headers[headerKey] = val
		}
	}

	data, err := json.Marshal(cached)
	if err != nil {
		log.Printf("WARNING: Cache middleware marshal error: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := cm.client.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("WARNING: Cache middleware Redis write error: %v", err)
	}
}

// responseCapture wraps gin.ResponseWriter to capture the response body
// while still writing it to the original writer.
type responseCapture struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// Write captures the response body while writing to the underlying writer.
func (rc *responseCapture) Write(data []byte) (int, error) {
	rc.body.Write(data)
	return rc.ResponseWriter.Write(data)
}

// WriteHeader captures the status code.
func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

// Status returns the captured status code, falling back to the underlying writer.
func (rc *responseCapture) Status() int {
	if rc.statusCode != 0 {
		return rc.statusCode
	}
	return rc.ResponseWriter.Status()
}
