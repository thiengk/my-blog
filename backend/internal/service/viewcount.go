package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key prefixes
	viewSeenKeyPrefix  = "view:seen:"  // view:seen:{slug}:{ip_hash} → "1" (TTL: 24h)
	viewBatchKeyPrefix = "view:batch:" // view:batch:{slug} → integer (batch counter)
	viewCountKeyPrefix = "view:count:" // view:count:{slug} → integer (cached count)

	// TTL durations
	viewSeenTTL  = 24 * time.Hour // Duplicate check TTL
	viewCountTTL = 5 * time.Minute // Cached count TTL
)

// ViewCountService defines the interface for view count operations.
type ViewCountService interface {
	// RecordView records a page view. Returns true if the view was counted
	// (not a duplicate), false if the same IP already viewed within 24h.
	RecordView(ctx context.Context, slug string, ip string) (bool, error)
	// GetCount returns the view count for a single post.
	// Checks cache first, falls back to DB.
	GetCount(ctx context.Context, slug string) (int64, error)
	// GetBulkCounts returns view counts for multiple slugs.
	GetBulkCounts(ctx context.Context, slugs []string) (map[string]int64, error)
	// FlushBatch flushes pending batch counts from Redis into PostgreSQL.
	FlushBatch(ctx context.Context) error
}

// viewCountService implements ViewCountService.
type viewCountService struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewViewCountService creates a new ViewCountService instance.
func NewViewCountService(db *pgxpool.Pool, redisClient *redis.Client) ViewCountService {
	return &viewCountService{
		db:    db,
		redis: redisClient,
	}
}

// hashIP creates a SHA-256 hash of the IP address for privacy.
func hashIP(ip string) string {
	h := sha256.Sum256([]byte(ip))
	return fmt.Sprintf("%x", h)
}

// RecordView checks if the IP has already viewed this post within 24h.
// If not, it increments the batch counter in Redis.
func (s *viewCountService) RecordView(ctx context.Context, slug string, ip string) (bool, error) {
	if s.redis == nil {
		return false, fmt.Errorf("redis client is not available")
	}

	ipHash := hashIP(ip)
	seenKey := fmt.Sprintf("%s%s:%s", viewSeenKeyPrefix, slug, ipHash)

	// Check if this IP already viewed this post within 24h
	exists, err := s.redis.Exists(ctx, seenKey).Result()
	if err != nil {
		log.Printf("WARNING: Redis error checking seen key: %v", err)
		// Fail-open: allow the view to be counted
	} else if exists > 0 {
		// Duplicate view within 24h, don't count
		return false, nil
	}

	// Mark this IP as having viewed the post (TTL 24h)
	err = s.redis.Set(ctx, seenKey, "1", viewSeenTTL).Err()
	if err != nil {
		log.Printf("WARNING: Redis error setting seen key: %v", err)
		// Continue anyway - worst case we count a duplicate later
	}

	// Increment the batch counter
	batchKey := fmt.Sprintf("%s%s", viewBatchKeyPrefix, slug)
	err = s.redis.Incr(ctx, batchKey).Err()
	if err != nil {
		log.Printf("WARNING: Redis error incrementing batch counter: %v", err)
		return false, fmt.Errorf("failed to increment batch counter: %w", err)
	}

	// Invalidate the cached count so next read gets fresh data
	countKey := fmt.Sprintf("%s%s", viewCountKeyPrefix, slug)
	_ = s.redis.Del(ctx, countKey).Err()

	return true, nil
}

// GetCount returns the view count for a single post.
// It checks the Redis cache first, then falls back to the database.
func (s *viewCountService) GetCount(ctx context.Context, slug string) (int64, error) {
	countKey := fmt.Sprintf("%s%s", viewCountKeyPrefix, slug)

	// Try to get from cache first
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, countKey).Int64()
		if err == nil {
			// Also add any pending batch count
			batchKey := fmt.Sprintf("%s%s", viewBatchKeyPrefix, slug)
			batchCount, batchErr := s.redis.Get(ctx, batchKey).Int64()
			if batchErr == nil {
				return cached + batchCount, nil
			}
			return cached, nil
		}
		// Cache miss or error - fall through to DB
		if err != redis.Nil {
			log.Printf("WARNING: Redis error getting cached count: %v", err)
		}
	}

	// Fallback to database
	if s.db == nil {
		return 0, fmt.Errorf("database is not available")
	}

	var count int64
	err := s.db.QueryRow(ctx,
		"SELECT view_count FROM post_views WHERE slug = $1", slug,
	).Scan(&count)
	if err != nil {
		// If no row found, return 0 (post hasn't been viewed yet)
		count = 0
	}

	// Cache the DB result
	if s.redis != nil {
		_ = s.redis.Set(ctx, countKey, count, viewCountTTL).Err()
	}

	// Add any pending batch count
	if s.redis != nil {
		batchKey := fmt.Sprintf("%s%s", viewBatchKeyPrefix, slug)
		batchCount, batchErr := s.redis.Get(ctx, batchKey).Int64()
		if batchErr == nil {
			count += batchCount
		}
	}

	return count, nil
}

// GetBulkCounts returns view counts for multiple slugs.
func (s *viewCountService) GetBulkCounts(ctx context.Context, slugs []string) (map[string]int64, error) {
	results := make(map[string]int64, len(slugs))

	for _, slug := range slugs {
		count, err := s.GetCount(ctx, slug)
		if err != nil {
			log.Printf("WARNING: Error getting count for slug %s: %v", slug, err)
			results[slug] = 0
			continue
		}
		results[slug] = count
	}

	return results, nil
}

// FlushBatch scans Redis for all view:batch:* keys and flushes their counts
// into PostgreSQL. It then clears the batch keys and updates the cache.
func (s *viewCountService) FlushBatch(ctx context.Context) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is not available")
	}
	if s.db == nil {
		return fmt.Errorf("database is not available")
	}

	// Scan for all batch keys
	pattern := fmt.Sprintf("%s*", viewBatchKeyPrefix)
	var cursor uint64
	var batchKeys []string

	for {
		keys, nextCursor, err := s.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan batch keys: %w", err)
		}
		batchKeys = append(batchKeys, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if len(batchKeys) == 0 {
		return nil
	}

	// Process each batch key
	for _, key := range batchKeys {
		// Extract slug from key: "view:batch:{slug}"
		slug := key[len(viewBatchKeyPrefix):]

		// Atomically get and delete the batch count
		count, err := s.redis.GetDel(ctx, key).Int64()
		if err != nil {
			if err == redis.Nil {
				continue // Key was already deleted (race condition)
			}
			log.Printf("WARNING: Failed to get batch count for %s: %v", slug, err)
			continue
		}

		if count <= 0 {
			continue
		}

		// Upsert into PostgreSQL
		_, err = s.db.Exec(ctx, `
			INSERT INTO post_views (slug, view_count, created_at, updated_at)
			VALUES ($1, $2, NOW(), NOW())
			ON CONFLICT (slug)
			DO UPDATE SET view_count = post_views.view_count + $2, updated_at = NOW()
		`, slug, count)
		if err != nil {
			// If DB write fails, put the count back in Redis
			log.Printf("ERROR: Failed to flush batch for slug %s: %v", slug, err)
			s.redis.IncrBy(ctx, key, count)
			continue
		}

		// Update the cache with the new total from DB
		var newTotal int64
		dbErr := s.db.QueryRow(ctx,
			"SELECT view_count FROM post_views WHERE slug = $1", slug,
		).Scan(&newTotal)
		if dbErr == nil {
			countKey := fmt.Sprintf("%s%s", viewCountKeyPrefix, slug)
			_ = s.redis.Set(ctx, countKey, newTotal, viewCountTTL).Err()
		}
	}

	return nil
}
