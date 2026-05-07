package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key prefixes for engagement
	likeSeenKeyPrefix  = "like:seen:"  // like:seen:{slug}:{ip_hash} → "1" (TTL: 24h)
	likeBatchKeyPrefix = "like:batch:" // like:batch:{slug} → integer (batch counter)

	shareSeenKeyPrefix  = "share:seen:"  // share:seen:{slug}:{ip_hash} → "1" (TTL: 24h)
	shareBatchKeyPrefix = "share:batch:" // share:batch:{slug} → integer (batch counter)

	engagementCountKeyPrefix = "engagement:count:" // engagement:count:{slug} → hash (likes, comments, shares)

	// TTL durations for engagement
	engagementSeenTTL  = 24 * time.Hour  // Duplicate like/share check TTL
	engagementCountTTL = 5 * time.Minute // Cached engagement count TTL
)

// EngagementService handles likes and shares with Redis batching.
type EngagementService interface {
	// RecordLike records a like for a post. Returns true if counted (not duplicate).
	RecordLike(ctx context.Context, slug string, ip string) (bool, error)
	// RecordShare records a share for a post with platform info. Returns true if counted (not duplicate).
	RecordShare(ctx context.Context, slug string, ip string, platform string) (bool, error)
	// GetCounts returns engagement counts (likes, comments, shares) for a single post.
	GetCounts(ctx context.Context, slug string) (*EngagementCounts, error)
	// GetBulkCounts returns engagement counts for multiple posts.
	GetBulkCounts(ctx context.Context, slugs []string) (map[string]*EngagementCounts, error)
	// FlushBatch flushes pending like/share counts from Redis to PostgreSQL.
	FlushBatch(ctx context.Context) error
}

// EngagementCounts holds the engagement metrics for a post.
type EngagementCounts struct {
	Likes    int64 `json:"likes"`
	Comments int64 `json:"comments"`
	Shares   int64 `json:"shares"`
}

// engagementService implements EngagementService.
type engagementService struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewEngagementService creates a new EngagementService instance.
func NewEngagementService(db *pgxpool.Pool, redisClient *redis.Client) EngagementService {
	return &engagementService{
		db:    db,
		redis: redisClient,
	}
}

// engagementHashIP creates a SHA-256 hash of the IP address for privacy.
// Reuses the same pattern as the view count service's hashIP.
func engagementHashIP(ip string) string {
	return hashIP(ip)
}

// RecordLike records a like for a post. Returns true if counted (not duplicate).
func (s *engagementService) RecordLike(ctx context.Context, slug string, ip string) (bool, error) {
	ipHash := engagementHashIP(ip)

	// Fallback: if Redis is unavailable, write directly to PostgreSQL
	if s.redis == nil {
		return s.recordLikeFallback(ctx, slug, ipHash)
	}

	seenKey := fmt.Sprintf("%s%s:%s", likeSeenKeyPrefix, slug, ipHash)

	// Check if this IP already liked this post within 24h
	exists, err := s.redis.Exists(ctx, seenKey).Result()
	if err != nil {
		log.Printf("WARNING: Redis error checking like seen key: %v", err)
		// Fallback to direct PostgreSQL write
		return s.recordLikeFallback(ctx, slug, ipHash)
	}
	if exists > 0 {
		// Duplicate like within 24h, don't count
		return false, nil
	}

	// Mark this IP as having liked the post (TTL 24h)
	err = s.redis.Set(ctx, seenKey, "1", engagementSeenTTL).Err()
	if err != nil {
		log.Printf("WARNING: Redis error setting like seen key: %v", err)
		// Fallback to direct PostgreSQL write
		return s.recordLikeFallback(ctx, slug, ipHash)
	}

	// Increment the batch counter
	batchKey := fmt.Sprintf("%s%s", likeBatchKeyPrefix, slug)
	err = s.redis.Incr(ctx, batchKey).Err()
	if err != nil {
		log.Printf("WARNING: Redis error incrementing like batch counter: %v", err)
		// Fallback to direct PostgreSQL write
		return s.recordLikeFallback(ctx, slug, ipHash)
	}

	// Invalidate the cached engagement count so next read gets fresh data
	countKey := fmt.Sprintf("%s%s", engagementCountKeyPrefix, slug)
	_ = s.redis.Del(ctx, countKey).Err()

	return true, nil
}

// recordLikeFallback writes a like directly to PostgreSQL when Redis is unavailable.
func (s *engagementService) recordLikeFallback(ctx context.Context, slug string, ipHash string) (bool, error) {
	if s.db == nil {
		return false, fmt.Errorf("database is not available")
	}

	// Upsert into post_engagement, incrementing like_count
	_, err := s.db.Exec(ctx, `
		INSERT INTO post_engagement (slug, like_count, comment_count, share_count, created_at, updated_at)
		VALUES ($1, 1, 0, 0, NOW(), NOW())
		ON CONFLICT (slug)
		DO UPDATE SET like_count = post_engagement.like_count + 1, updated_at = NOW()
	`, slug)
	if err != nil {
		return false, fmt.Errorf("failed to record like in PostgreSQL: %w", err)
	}

	return true, nil
}

// RecordShare records a share for a post with platform info. Returns true if counted (not duplicate).
func (s *engagementService) RecordShare(ctx context.Context, slug string, ip string, platform string) (bool, error) {
	ipHash := engagementHashIP(ip)

	// Fallback: if Redis is unavailable, write directly to PostgreSQL
	if s.redis == nil {
		return s.recordShareFallback(ctx, slug, ipHash, platform)
	}

	seenKey := fmt.Sprintf("%s%s:%s", shareSeenKeyPrefix, slug, ipHash)

	// Check if this IP already shared this post within 24h
	exists, err := s.redis.Exists(ctx, seenKey).Result()
	if err != nil {
		log.Printf("WARNING: Redis error checking share seen key: %v", err)
		// Fallback to direct PostgreSQL write
		return s.recordShareFallback(ctx, slug, ipHash, platform)
	}
	if exists > 0 {
		// Duplicate share within 24h, don't count
		return false, nil
	}

	// Mark this IP as having shared the post (TTL 24h)
	err = s.redis.Set(ctx, seenKey, "1", engagementSeenTTL).Err()
	if err != nil {
		log.Printf("WARNING: Redis error setting share seen key: %v", err)
		// Fallback to direct PostgreSQL write
		return s.recordShareFallback(ctx, slug, ipHash, platform)
	}

	// Increment the batch counter
	batchKey := fmt.Sprintf("%s%s", shareBatchKeyPrefix, slug)
	err = s.redis.Incr(ctx, batchKey).Err()
	if err != nil {
		log.Printf("WARNING: Redis error incrementing share batch counter: %v", err)
		// Fallback to direct PostgreSQL write
		return s.recordShareFallback(ctx, slug, ipHash, platform)
	}

	// Insert into share_logs for platform tracking
	if s.db != nil {
		_, err = s.db.Exec(ctx, `
			INSERT INTO share_logs (slug, platform, ip_hash, shared_at)
			VALUES ($1, $2, $3, NOW())
		`, slug, platform, ipHash)
		if err != nil {
			log.Printf("WARNING: Failed to insert share log: %v", err)
			// Non-critical: share count is still tracked via Redis batch
		}
	}

	// Invalidate the cached engagement count so next read gets fresh data
	countKey := fmt.Sprintf("%s%s", engagementCountKeyPrefix, slug)
	_ = s.redis.Del(ctx, countKey).Err()

	return true, nil
}

// recordShareFallback writes a share directly to PostgreSQL when Redis is unavailable.
func (s *engagementService) recordShareFallback(ctx context.Context, slug string, ipHash string, platform string) (bool, error) {
	if s.db == nil {
		return false, fmt.Errorf("database is not available")
	}

	// Upsert into post_engagement, incrementing share_count
	_, err := s.db.Exec(ctx, `
		INSERT INTO post_engagement (slug, like_count, comment_count, share_count, created_at, updated_at)
		VALUES ($1, 0, 0, 1, NOW(), NOW())
		ON CONFLICT (slug)
		DO UPDATE SET share_count = post_engagement.share_count + 1, updated_at = NOW()
	`, slug)
	if err != nil {
		return false, fmt.Errorf("failed to record share in PostgreSQL: %w", err)
	}

	// Insert into share_logs for platform tracking
	_, err = s.db.Exec(ctx, `
		INSERT INTO share_logs (slug, platform, ip_hash, shared_at)
		VALUES ($1, $2, $3, NOW())
	`, slug, platform, ipHash)
	if err != nil {
		log.Printf("WARNING: Failed to insert share log in fallback: %v", err)
		// Non-critical: share count is already tracked
	}

	return true, nil
}

// GetCounts returns engagement counts for a single post.
// Checks Redis hash cache first, falls back to PostgreSQL, caches result with 5-min TTL,
// and adds any pending batch counts from Redis.
func (s *engagementService) GetCounts(ctx context.Context, slug string) (*EngagementCounts, error) {
	countKey := fmt.Sprintf("%s%s", engagementCountKeyPrefix, slug)

	// Try to get from Redis hash cache first
	if s.redis != nil {
		cached, err := s.redis.HGetAll(ctx, countKey).Result()
		if err == nil && len(cached) > 0 {
			counts := &EngagementCounts{}
			if v, ok := cached["likes"]; ok {
				counts.Likes, _ = strconv.ParseInt(v, 10, 64)
			}
			if v, ok := cached["comments"]; ok {
				counts.Comments, _ = strconv.ParseInt(v, 10, 64)
			}
			if v, ok := cached["shares"]; ok {
				counts.Shares, _ = strconv.ParseInt(v, 10, 64)
			}

			// Add pending batch counts
			s.addPendingBatchCounts(ctx, slug, counts)
			return counts, nil
		}
		// Cache miss or error - fall through to DB
		if err != nil {
			log.Printf("WARNING: Redis error getting cached engagement counts: %v", err)
		}
	}

	// Fallback to database
	if s.db == nil {
		return &EngagementCounts{}, nil
	}

	counts := &EngagementCounts{}
	err := s.db.QueryRow(ctx,
		"SELECT like_count, comment_count, share_count FROM post_engagement WHERE slug = $1", slug,
	).Scan(&counts.Likes, &counts.Comments, &counts.Shares)
	if err != nil {
		// If no row found, return zeros (post hasn't had engagement yet)
		counts = &EngagementCounts{}
	}

	// Cache the DB result in Redis hash with TTL
	if s.redis != nil {
		pipe := s.redis.Pipeline()
		pipe.HSet(ctx, countKey, map[string]interface{}{
			"likes":    counts.Likes,
			"comments": counts.Comments,
			"shares":   counts.Shares,
		})
		pipe.Expire(ctx, countKey, engagementCountTTL)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("WARNING: Redis error caching engagement counts: %v", err)
		}
	}

	// Add pending batch counts
	s.addPendingBatchCounts(ctx, slug, counts)

	return counts, nil
}

// GetBulkCounts returns engagement counts for multiple posts.
// Limits to a maximum of 50 slugs per request.
func (s *engagementService) GetBulkCounts(ctx context.Context, slugs []string) (map[string]*EngagementCounts, error) {
	// Enforce maximum of 50 slugs
	if len(slugs) > 50 {
		slugs = slugs[:50]
	}

	results := make(map[string]*EngagementCounts, len(slugs))

	for _, slug := range slugs {
		counts, err := s.GetCounts(ctx, slug)
		if err != nil {
			log.Printf("WARNING: Error getting engagement counts for slug %s: %v", slug, err)
			results[slug] = &EngagementCounts{}
			continue
		}
		results[slug] = counts
	}

	return results, nil
}

// addPendingBatchCounts adds any pending like/share batch counts from Redis to the counts.
func (s *engagementService) addPendingBatchCounts(ctx context.Context, slug string, counts *EngagementCounts) {
	if s.redis == nil {
		return
	}

	// Add pending like batch count
	likeBatchKey := fmt.Sprintf("%s%s", likeBatchKeyPrefix, slug)
	likeBatch, err := s.redis.Get(ctx, likeBatchKey).Int64()
	if err == nil {
		counts.Likes += likeBatch
	}

	// Add pending share batch count
	shareBatchKey := fmt.Sprintf("%s%s", shareBatchKeyPrefix, slug)
	shareBatch, err := s.redis.Get(ctx, shareBatchKey).Int64()
	if err == nil {
		counts.Shares += shareBatch
	}
}

// FlushBatch flushes pending like/share counts from Redis to PostgreSQL.
// It scans for like:batch:* and share:batch:* keys, atomically gets and deletes them,
// then upserts the counts into the post_engagement table.
// If a DB write fails, the count is put back into Redis for retry on the next cycle.
func (s *engagementService) FlushBatch(ctx context.Context) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is not available")
	}
	if s.db == nil {
		return fmt.Errorf("database is not available")
	}

	// Flush like batch keys
	if err := s.flushBatchKeys(ctx, likeBatchKeyPrefix, "like_count"); err != nil {
		log.Printf("ERROR: Failed to flush like batch keys: %v", err)
	}

	// Flush share batch keys
	if err := s.flushBatchKeys(ctx, shareBatchKeyPrefix, "share_count"); err != nil {
		log.Printf("ERROR: Failed to flush share batch keys: %v", err)
	}

	return nil
}

// flushBatchKeys scans Redis for keys matching the given prefix, atomically gets and deletes
// each key's value, and upserts the count into the specified column of post_engagement.
func (s *engagementService) flushBatchKeys(ctx context.Context, keyPrefix string, column string) error {
	// Scan for all batch keys matching the prefix
	pattern := fmt.Sprintf("%s*", keyPrefix)
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
		// Extract slug from key: "{prefix}{slug}"
		slug := key[len(keyPrefix):]

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

		// Build initial values for the INSERT based on which column we're flushing
		var likeVal, shareVal int64
		if column == "like_count" {
			likeVal = count
		} else {
			shareVal = count
		}

		// Upsert into PostgreSQL post_engagement table
		query := fmt.Sprintf(`
			INSERT INTO post_engagement (slug, like_count, comment_count, share_count, created_at, updated_at)
			VALUES ($1, $2, 0, $3, NOW(), NOW())
			ON CONFLICT (slug)
			DO UPDATE SET %s = post_engagement.%s + $4, updated_at = NOW()
		`, column, column)

		_, err = s.db.Exec(ctx, query, slug, likeVal, shareVal, count)
		if err != nil {
			// If DB write fails, put the count back in Redis for retry on next cycle
			log.Printf("ERROR: Failed to flush %s batch for slug %s: %v", column, slug, err)
			s.redis.IncrBy(ctx, key, count)
			continue
		}

		// Invalidate the cached engagement count so next read gets fresh data
		countKey := fmt.Sprintf("%s%s", engagementCountKeyPrefix, slug)
		_ = s.redis.Del(ctx, countKey).Err()
	}

	return nil
}
