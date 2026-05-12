package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key for cached recommendations
	recommendationsTopKey = "recommendations:top"

	// Default configuration values
	defaultLikeWeight    = 1.0
	defaultCommentWeight = 2.0
	defaultShareWeight   = 3.0
	defaultCacheTTL      = 5 * time.Minute
	defaultMaxResults    = 50
)

// RecommendationService calculates and caches post rankings based on engagement score.
type RecommendationService interface {
	// GetTopPosts returns the top N posts sorted by engagement score.
	GetTopPosts(ctx context.Context, limit int) ([]*RankedPost, error)
	// RecalculateRankings forces a recalculation of rankings.
	RecalculateRankings(ctx context.Context) error
}

// RankedPost represents a post with its calculated engagement score.
type RankedPost struct {
	Slug            string  `json:"slug"`
	EngagementScore float64 `json:"engagement_score"`
	Likes           int64   `json:"likes"`
	Comments        int64   `json:"comments"`
	Shares          int64   `json:"shares"`
}

// RecommendationConfig holds configurable weights and settings for the recommendation engine.
type RecommendationConfig struct {
	LikeWeight    float64       // Weight for likes in score calculation (default: 1.0)
	CommentWeight float64       // Weight for comments in score calculation (default: 2.0)
	ShareWeight   float64       // Weight for shares in score calculation (default: 3.0)
	CacheTTL      time.Duration // TTL for cached rankings (default: 5 minutes)
	MaxResults    int           // Maximum number of results to return (default: 50)
}

// recommendationService implements RecommendationService.
type recommendationService struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	config RecommendationConfig
}

// NewRecommendationService creates a new RecommendationService instance with default configuration.
func NewRecommendationService(db *pgxpool.Pool, redisClient *redis.Client) RecommendationService {
	return NewRecommendationServiceWithConfig(db, redisClient, RecommendationConfig{})
}

// NewRecommendationServiceWithConfig creates a new RecommendationService instance with custom configuration.
// Zero-value fields in the config will be replaced with defaults.
func NewRecommendationServiceWithConfig(db *pgxpool.Pool, redisClient *redis.Client, config RecommendationConfig) RecommendationService {
	if config.LikeWeight == 0 {
		config.LikeWeight = defaultLikeWeight
	}
	if config.CommentWeight == 0 {
		config.CommentWeight = defaultCommentWeight
	}
	if config.ShareWeight == 0 {
		config.ShareWeight = defaultShareWeight
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = defaultCacheTTL
	}
	if config.MaxResults == 0 {
		config.MaxResults = defaultMaxResults
	}

	return &recommendationService{
		db:     db,
		redis:  redisClient,
		config: config,
	}
}

// GetTopPosts returns the top N posts sorted by engagement score.
// It checks the Redis cache first; on cache miss it triggers RecalculateRankings.
func (s *recommendationService) GetTopPosts(ctx context.Context, limit int) ([]*RankedPost, error) {
	// Clamp limit to valid range
	if limit <= 0 {
		limit = 10
	}
	if limit > s.config.MaxResults {
		limit = s.config.MaxResults
	}

	// Try to read from Redis cache
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, recommendationsTopKey).Bytes()
		if err == nil && len(cached) > 0 {
			// Cache hit — unmarshal the JSON list
			var posts []*RankedPost
			if jsonErr := json.Unmarshal(cached, &posts); jsonErr == nil {
				// Return up to the requested limit
				if len(posts) > limit {
					posts = posts[:limit]
				}
				return posts, nil
			}
			// If unmarshal fails, fall through to recalculate
			log.Printf("[recommendation] failed to unmarshal cached rankings: %v", err)
		}
	}

	// Cache miss or Redis unavailable — recalculate rankings
	if err := s.RecalculateRankings(ctx); err != nil {
		// If recalculation fails, return empty list instead of error
		log.Printf("[recommendation] failed to recalculate rankings: %v", err)
		return []*RankedPost{}, nil
	}

	// Read the freshly cached data
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, recommendationsTopKey).Bytes()
		if err == nil {
			var posts []*RankedPost
			if err := json.Unmarshal(cached, &posts); err == nil {
				if len(posts) > limit {
					posts = posts[:limit]
				}
				return posts, nil
			}
		}
	}

	// If we still can't read from cache, query DB directly
	return s.queryRankingsFromDB(ctx, limit)
}

// RecalculateRankings queries PostgreSQL to compute engagement scores and caches the result in Redis.
// Score formula: (like_count × like_weight) + (comment_count × comment_weight) + (share_count × share_weight)
// Results are sorted by score DESC, with created_at DESC as tiebreaker.
func (s *recommendationService) RecalculateRankings(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database is not available")
	}

	query := `
		SELECT slug, like_count, comment_count, share_count,
			(like_count * $1 + comment_count * $2 + share_count * $3) AS engagement_score
		FROM post_engagement
		ORDER BY engagement_score DESC, created_at DESC
		LIMIT $4
	`

	rows, err := s.db.Query(ctx, query,
		s.config.LikeWeight,
		s.config.CommentWeight,
		s.config.ShareWeight,
		s.config.MaxResults,
	)
	if err != nil {
		return fmt.Errorf("failed to query engagement rankings: %w", err)
	}
	defer rows.Close()

	var posts []*RankedPost
	for rows.Next() {
		var p RankedPost
		if err := rows.Scan(&p.Slug, &p.Likes, &p.Comments, &p.Shares, &p.EngagementScore); err != nil {
			return fmt.Errorf("failed to scan ranked post row: %w", err)
		}
		posts = append(posts, &p)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating ranked post rows: %w", err)
	}

	// If no posts found, store an empty array
	if posts == nil {
		posts = []*RankedPost{}
	}

	// Serialize and cache in Redis with TTL
	if s.redis != nil {
		data, err := json.Marshal(posts)
		if err != nil {
			return fmt.Errorf("failed to marshal rankings: %w", err)
		}

		if err := s.redis.Set(ctx, recommendationsTopKey, data, s.config.CacheTTL).Err(); err != nil {
			log.Printf("[recommendation] failed to cache rankings in Redis: %v", err)
			// Non-fatal: we still have the data, just won't be cached
		}
	}

	return nil
}

// queryRankingsFromDB queries rankings directly from DB without caching.
// Used as fallback when Redis is unavailable.
func (s *recommendationService) queryRankingsFromDB(ctx context.Context, limit int) ([]*RankedPost, error) {
	if s.db == nil {
		return []*RankedPost{}, nil
	}

	query := `
		SELECT slug, like_count, comment_count, share_count,
			(like_count * $1 + comment_count * $2 + share_count * $3) AS engagement_score
		FROM post_engagement
		ORDER BY engagement_score DESC, created_at DESC
		LIMIT $4
	`

	rows, err := s.db.Query(ctx, query,
		s.config.LikeWeight,
		s.config.CommentWeight,
		s.config.ShareWeight,
		limit,
	)
	if err != nil {
		log.Printf("[recommendation] failed to query DB directly: %v", err)
		return []*RankedPost{}, nil
	}
	defer rows.Close()

	var posts []*RankedPost
	for rows.Next() {
		var p RankedPost
		if err := rows.Scan(&p.Slug, &p.Likes, &p.Comments, &p.Shares, &p.EngagementScore); err != nil {
			log.Printf("[recommendation] failed to scan row: %v", err)
			continue
		}
		posts = append(posts, &p)
	}

	if posts == nil {
		posts = []*RankedPost{}
	}
	return posts, nil
}
