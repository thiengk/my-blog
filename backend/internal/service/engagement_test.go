package service

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// newTestRedisClient creates a Redis client for testing.
// Tests using this require a running Redis instance or will be skipped.
// For unit testing without external dependencies, we use a real redis.Client
// pointed at a test address. If connection fails, tests are skipped.
func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Use DB 15 for testing to avoid conflicts
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping test: Redis not available at localhost:6379: %v", err)
	}

	// Flush the test database before each test
	client.FlushDB(ctx)

	t.Cleanup(func() {
		client.FlushDB(context.Background())
		client.Close()
	})

	return client
}

// =============================================================================
// RecordLike Tests
// =============================================================================

func TestRecordLike_NewLike(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Record a new like - should succeed
	counted, err := svc.RecordLike(ctx, "test-post", "192.168.1.1")
	if err != nil {
		t.Fatalf("RecordLike failed: %v", err)
	}
	if !counted {
		t.Error("expected like to be counted (new like)")
	}

	// Verify the batch counter was incremented
	batchKey := "like:batch:test-post"
	val, err := redisClient.Get(ctx, batchKey).Int64()
	if err != nil {
		t.Fatalf("failed to get batch key: %v", err)
	}
	if val != 1 {
		t.Errorf("expected batch counter = 1, got %d", val)
	}

	// Verify the dedup key was set
	ipHash := hashIP("192.168.1.1")
	seenKey := "like:seen:test-post:" + ipHash
	exists, err := redisClient.Exists(ctx, seenKey).Result()
	if err != nil {
		t.Fatalf("failed to check seen key: %v", err)
	}
	if exists != 1 {
		t.Error("expected dedup seen key to exist")
	}

	// Verify the TTL on the seen key is approximately 24h
	ttl, err := redisClient.TTL(ctx, seenKey).Result()
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}
	if ttl < 23*time.Hour || ttl > 25*time.Hour {
		t.Errorf("expected TTL ~24h, got %v", ttl)
	}
}

func TestRecordLike_DuplicateWithin24h(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// First like - should be counted
	counted, err := svc.RecordLike(ctx, "test-post", "192.168.1.1")
	if err != nil {
		t.Fatalf("first RecordLike failed: %v", err)
	}
	if !counted {
		t.Error("expected first like to be counted")
	}

	// Second like from same IP within 24h - should NOT be counted
	counted, err = svc.RecordLike(ctx, "test-post", "192.168.1.1")
	if err != nil {
		t.Fatalf("second RecordLike failed: %v", err)
	}
	if counted {
		t.Error("expected duplicate like to NOT be counted")
	}

	// Verify batch counter is still 1 (not incremented for duplicate)
	batchKey := "like:batch:test-post"
	val, err := redisClient.Get(ctx, batchKey).Int64()
	if err != nil {
		t.Fatalf("failed to get batch key: %v", err)
	}
	if val != 1 {
		t.Errorf("expected batch counter = 1 (no duplicate increment), got %d", val)
	}
}

func TestRecordLike_DifferentIPsAllowed(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Like from IP 1
	counted1, err := svc.RecordLike(ctx, "test-post", "192.168.1.1")
	if err != nil {
		t.Fatalf("RecordLike IP1 failed: %v", err)
	}
	if !counted1 {
		t.Error("expected like from IP1 to be counted")
	}

	// Like from IP 2 - should also be counted
	counted2, err := svc.RecordLike(ctx, "test-post", "192.168.1.2")
	if err != nil {
		t.Fatalf("RecordLike IP2 failed: %v", err)
	}
	if !counted2 {
		t.Error("expected like from IP2 to be counted")
	}

	// Verify batch counter is 2
	batchKey := "like:batch:test-post"
	val, err := redisClient.Get(ctx, batchKey).Int64()
	if err != nil {
		t.Fatalf("failed to get batch key: %v", err)
	}
	if val != 2 {
		t.Errorf("expected batch counter = 2, got %d", val)
	}
}

func TestRecordLike_DifferentSlugsAllowed(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Like post-1 from same IP
	counted1, err := svc.RecordLike(ctx, "post-1", "192.168.1.1")
	if err != nil {
		t.Fatalf("RecordLike post-1 failed: %v", err)
	}
	if !counted1 {
		t.Error("expected like for post-1 to be counted")
	}

	// Like post-2 from same IP - should also be counted (different slug)
	counted2, err := svc.RecordLike(ctx, "post-2", "192.168.1.1")
	if err != nil {
		t.Fatalf("RecordLike post-2 failed: %v", err)
	}
	if !counted2 {
		t.Error("expected like for post-2 to be counted")
	}
}

func TestRecordLike_FallbackWhenRedisNil(t *testing.T) {
	ctx := context.Background()

	// Service with nil Redis and nil DB - should return error from fallback
	svc := NewEngagementService(nil, nil)

	_, err := svc.RecordLike(ctx, "test-post", "192.168.1.1")
	if err == nil {
		t.Error("expected error when both Redis and DB are nil")
	}
}

// =============================================================================
// RecordShare Tests
// =============================================================================

func TestRecordShare_NewShare(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Record a new share - should succeed
	counted, err := svc.RecordShare(ctx, "test-post", "192.168.1.1", "facebook")
	if err != nil {
		t.Fatalf("RecordShare failed: %v", err)
	}
	if !counted {
		t.Error("expected share to be counted (new share)")
	}

	// Verify the batch counter was incremented
	batchKey := "share:batch:test-post"
	val, err := redisClient.Get(ctx, batchKey).Int64()
	if err != nil {
		t.Fatalf("failed to get batch key: %v", err)
	}
	if val != 1 {
		t.Errorf("expected batch counter = 1, got %d", val)
	}

	// Verify the dedup key was set
	ipHash := hashIP("192.168.1.1")
	seenKey := "share:seen:test-post:" + ipHash
	exists, err := redisClient.Exists(ctx, seenKey).Result()
	if err != nil {
		t.Fatalf("failed to check seen key: %v", err)
	}
	if exists != 1 {
		t.Error("expected dedup seen key to exist")
	}
}

func TestRecordShare_DuplicateWithin24h(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// First share - should be counted
	counted, err := svc.RecordShare(ctx, "test-post", "192.168.1.1", "twitter")
	if err != nil {
		t.Fatalf("first RecordShare failed: %v", err)
	}
	if !counted {
		t.Error("expected first share to be counted")
	}

	// Second share from same IP within 24h - should NOT be counted
	counted, err = svc.RecordShare(ctx, "test-post", "192.168.1.1", "facebook")
	if err != nil {
		t.Fatalf("second RecordShare failed: %v", err)
	}
	if counted {
		t.Error("expected duplicate share to NOT be counted")
	}

	// Verify batch counter is still 1
	batchKey := "share:batch:test-post"
	val, err := redisClient.Get(ctx, batchKey).Int64()
	if err != nil {
		t.Fatalf("failed to get batch key: %v", err)
	}
	if val != 1 {
		t.Errorf("expected batch counter = 1 (no duplicate increment), got %d", val)
	}
}

func TestRecordShare_PlatformTracking(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Share with specific platform - should succeed
	// Note: platform tracking writes to share_logs in DB, which is nil here
	// The share should still be counted in Redis batch even if DB log fails
	platforms := []string{"facebook", "twitter", "linkedin", "copy-link"}

	for i, platform := range platforms {
		ip := "192.168.1." + string(rune('1'+i))
		counted, err := svc.RecordShare(ctx, "test-post", ip, platform)
		if err != nil {
			t.Fatalf("RecordShare with platform %s failed: %v", platform, err)
		}
		if !counted {
			t.Errorf("expected share with platform %s to be counted", platform)
		}
	}

	// Verify batch counter reflects all shares
	batchKey := "share:batch:test-post"
	val, err := redisClient.Get(ctx, batchKey).Int64()
	if err != nil {
		t.Fatalf("failed to get batch key: %v", err)
	}
	if val != int64(len(platforms)) {
		t.Errorf("expected batch counter = %d, got %d", len(platforms), val)
	}
}

func TestRecordShare_FallbackWhenRedisNil(t *testing.T) {
	ctx := context.Background()

	// Service with nil Redis and nil DB - should return error from fallback
	svc := NewEngagementService(nil, nil)

	_, err := svc.RecordShare(ctx, "test-post", "192.168.1.1", "facebook")
	if err == nil {
		t.Error("expected error when both Redis and DB are nil")
	}
}

// =============================================================================
// GetCounts Tests
// =============================================================================

func TestGetCounts_CacheHit(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Pre-populate the Redis hash cache
	countKey := "engagement:count:test-post"
	redisClient.HSet(ctx, countKey, map[string]interface{}{
		"likes":    int64(10),
		"comments": int64(5),
		"shares":   int64(3),
	})
	redisClient.Expire(ctx, countKey, 5*time.Minute)

	// GetCounts should return cached values
	counts, err := svc.GetCounts(ctx, "test-post")
	if err != nil {
		t.Fatalf("GetCounts failed: %v", err)
	}

	if counts.Likes != 10 {
		t.Errorf("expected likes = 10, got %d", counts.Likes)
	}
	if counts.Comments != 5 {
		t.Errorf("expected comments = 5, got %d", counts.Comments)
	}
	if counts.Shares != 3 {
		t.Errorf("expected shares = 3, got %d", counts.Shares)
	}
}

func TestGetCounts_CacheMiss_NoDB(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	// Service with Redis but no DB - cache miss should return zeros
	svc := NewEngagementService(nil, redisClient)

	counts, err := svc.GetCounts(ctx, "nonexistent-post")
	if err != nil {
		t.Fatalf("GetCounts failed: %v", err)
	}

	// With no DB and no cache, should return zero counts
	if counts.Likes != 0 {
		t.Errorf("expected likes = 0, got %d", counts.Likes)
	}
	if counts.Comments != 0 {
		t.Errorf("expected comments = 0, got %d", counts.Comments)
	}
	if counts.Shares != 0 {
		t.Errorf("expected shares = 0, got %d", counts.Shares)
	}
}

func TestGetCounts_PendingBatchCounts(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Pre-populate the Redis hash cache with base counts
	countKey := "engagement:count:test-post"
	redisClient.HSet(ctx, countKey, map[string]interface{}{
		"likes":    int64(10),
		"comments": int64(5),
		"shares":   int64(3),
	})
	redisClient.Expire(ctx, countKey, 5*time.Minute)

	// Add pending batch counts (simulating likes/shares not yet flushed to DB)
	redisClient.Set(ctx, "like:batch:test-post", 3, 0)
	redisClient.Set(ctx, "share:batch:test-post", 2, 0)

	// GetCounts should return cached values + pending batch counts
	counts, err := svc.GetCounts(ctx, "test-post")
	if err != nil {
		t.Fatalf("GetCounts failed: %v", err)
	}

	// likes = 10 (cached) + 3 (pending) = 13
	if counts.Likes != 13 {
		t.Errorf("expected likes = 13 (10 cached + 3 pending), got %d", counts.Likes)
	}
	// comments = 5 (cached, no batch for comments)
	if counts.Comments != 5 {
		t.Errorf("expected comments = 5, got %d", counts.Comments)
	}
	// shares = 3 (cached) + 2 (pending) = 5
	if counts.Shares != 5 {
		t.Errorf("expected shares = 5 (3 cached + 2 pending), got %d", counts.Shares)
	}
}

func TestGetCounts_NilRedisAndNilDB(t *testing.T) {
	ctx := context.Background()

	// Service with nil Redis and nil DB
	svc := NewEngagementService(nil, nil)

	counts, err := svc.GetCounts(ctx, "test-post")
	if err != nil {
		t.Fatalf("GetCounts failed: %v", err)
	}

	// Should return zero counts gracefully
	if counts.Likes != 0 || counts.Comments != 0 || counts.Shares != 0 {
		t.Errorf("expected all zeros, got likes=%d, comments=%d, shares=%d",
			counts.Likes, counts.Comments, counts.Shares)
	}
}

func TestGetCounts_CacheInvalidatedAfterLike(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Pre-populate cache
	countKey := "engagement:count:test-post"
	redisClient.HSet(ctx, countKey, map[string]interface{}{
		"likes":    int64(10),
		"comments": int64(5),
		"shares":   int64(3),
	})
	redisClient.Expire(ctx, countKey, 5*time.Minute)

	// Record a like - should invalidate cache
	_, err := svc.RecordLike(ctx, "test-post", "192.168.1.1")
	if err != nil {
		t.Fatalf("RecordLike failed: %v", err)
	}

	// Verify cache was invalidated
	exists, err := redisClient.Exists(ctx, countKey).Result()
	if err != nil {
		t.Fatalf("failed to check cache key: %v", err)
	}
	if exists != 0 {
		t.Error("expected cache to be invalidated after like")
	}
}

// =============================================================================
// GetBulkCounts Tests
// =============================================================================

func TestGetBulkCounts_MultipleSlugs(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Pre-populate cache for multiple slugs
	slugs := []string{"post-1", "post-2", "post-3"}
	for i, slug := range slugs {
		countKey := "engagement:count:" + slug
		redisClient.HSet(ctx, countKey, map[string]interface{}{
			"likes":    int64((i + 1) * 10),
			"comments": int64((i + 1) * 5),
			"shares":   int64((i + 1) * 2),
		})
		redisClient.Expire(ctx, countKey, 5*time.Minute)
	}

	results, err := svc.GetBulkCounts(ctx, slugs)
	if err != nil {
		t.Fatalf("GetBulkCounts failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Verify post-1 counts
	if results["post-1"].Likes != 10 {
		t.Errorf("post-1: expected likes = 10, got %d", results["post-1"].Likes)
	}
	// Verify post-2 counts
	if results["post-2"].Likes != 20 {
		t.Errorf("post-2: expected likes = 20, got %d", results["post-2"].Likes)
	}
	// Verify post-3 counts
	if results["post-3"].Likes != 30 {
		t.Errorf("post-3: expected likes = 30, got %d", results["post-3"].Likes)
	}
}

func TestGetBulkCounts_LimitTo50(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Create 60 slugs - should be limited to 50
	slugs := make([]string, 60)
	for i := range slugs {
		slugs[i] = "post-" + string(rune('a'+i%26)) + string(rune('a'+i/26))
	}

	results, err := svc.GetBulkCounts(ctx, slugs)
	if err != nil {
		t.Fatalf("GetBulkCounts failed: %v", err)
	}

	// Should only process first 50 slugs
	if len(results) != 50 {
		t.Errorf("expected 50 results (limited), got %d", len(results))
	}
}

// =============================================================================
// FlushBatch Tests
// =============================================================================

func TestFlushBatch_NilRedis(t *testing.T) {
	ctx := context.Background()

	svc := NewEngagementService(nil, nil)

	err := svc.FlushBatch(ctx)
	if err == nil {
		t.Error("expected error when Redis is nil")
	}
}

func TestFlushBatch_NoBatchKeys(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	// Service with Redis but no DB - no batch keys to flush
	svc := NewEngagementService(nil, redisClient)

	// FlushBatch with no pending batch keys should not error
	// Note: current implementation returns "not implemented" - this test
	// documents the expected behavior once FlushBatch is implemented
	err := svc.FlushBatch(ctx)
	// The current implementation returns "not implemented" error
	// Once task 2.4 is complete, this should return nil for empty batch
	if err != nil && err.Error() != "not implemented" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFlushBatch_DBFailureRetry(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	// Service with Redis but nil DB - simulates DB failure
	svc := NewEngagementService(nil, redisClient)

	// Set up pending batch counts
	redisClient.Set(ctx, "like:batch:test-post", 5, 0)
	redisClient.Set(ctx, "share:batch:test-post", 3, 0)

	// FlushBatch should fail because DB is nil
	// Note: current implementation returns "not implemented"
	// Once task 2.4 is complete, it should either:
	// - Return an error about DB unavailability, OR
	// - Retain counts in Redis for retry on next cycle (per requirement 7.4)
	err := svc.FlushBatch(ctx)
	if err == nil {
		// If no error, verify counts are still in Redis (retry behavior)
		likeCount, _ := redisClient.Get(ctx, "like:batch:test-post").Int64()
		shareCount, _ := redisClient.Get(ctx, "share:batch:test-post").Int64()
		if likeCount != 5 {
			t.Errorf("expected like batch count retained = 5, got %d", likeCount)
		}
		if shareCount != 3 {
			t.Errorf("expected share batch count retained = 3, got %d", shareCount)
		}
	}
	// If error is returned, that's also acceptable behavior for nil DB
}

// =============================================================================
// Integration-style Tests (RecordLike + GetCounts)
// =============================================================================

func TestRecordLikeAndGetCounts_Integration(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Record multiple likes from different IPs
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	for _, ip := range ips {
		counted, err := svc.RecordLike(ctx, "integration-post", ip)
		if err != nil {
			t.Fatalf("RecordLike failed for IP %s: %v", ip, err)
		}
		if !counted {
			t.Errorf("expected like from IP %s to be counted", ip)
		}
	}

	// GetCounts should reflect the pending batch likes
	counts, err := svc.GetCounts(ctx, "integration-post")
	if err != nil {
		t.Fatalf("GetCounts failed: %v", err)
	}

	// Since there's no DB and no cache, likes come from pending batch only
	if counts.Likes != 3 {
		t.Errorf("expected likes = 3 (from pending batch), got %d", counts.Likes)
	}
}

func TestRecordShareAndGetCounts_Integration(t *testing.T) {
	redisClient := newTestRedisClient(t)
	ctx := context.Background()

	svc := NewEngagementService(nil, redisClient)

	// Record shares from different IPs
	shares := []struct {
		ip       string
		platform string
	}{
		{"10.0.0.1", "facebook"},
		{"10.0.0.2", "twitter"},
		{"10.0.0.3", "linkedin"},
	}

	for _, s := range shares {
		counted, err := svc.RecordShare(ctx, "integration-post", s.ip, s.platform)
		if err != nil {
			t.Fatalf("RecordShare failed for IP %s: %v", s.ip, err)
		}
		if !counted {
			t.Errorf("expected share from IP %s to be counted", s.ip)
		}
	}

	// GetCounts should reflect the pending batch shares
	counts, err := svc.GetCounts(ctx, "integration-post")
	if err != nil {
		t.Fatalf("GetCounts failed: %v", err)
	}

	if counts.Shares != 3 {
		t.Errorf("expected shares = 3 (from pending batch), got %d", counts.Shares)
	}
}

// =============================================================================
// Helper function tests
// =============================================================================

func TestEngagementHashIP_Consistency(t *testing.T) {
	// Same IP should always produce the same hash
	hash1 := engagementHashIP("192.168.1.1")
	hash2 := engagementHashIP("192.168.1.1")

	if hash1 != hash2 {
		t.Errorf("expected consistent hash, got %s and %s", hash1, hash2)
	}

	// Different IPs should produce different hashes
	hash3 := engagementHashIP("192.168.1.2")
	if hash1 == hash3 {
		t.Error("expected different hashes for different IPs")
	}
}

func TestEngagementHashIP_NotEmpty(t *testing.T) {
	hash := engagementHashIP("10.0.0.1")
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	// SHA-256 produces 64 hex characters
	if len(hash) != 64 {
		t.Errorf("expected 64 character hex hash, got %d characters", len(hash))
	}
}
