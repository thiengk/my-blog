package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key prefix for cached comment count
	commentCountKeyPrefix = "comment:count:" // comment:count:{slug} → integer

	// TTL for cached comment count
	commentCountTTL = 5 * time.Minute

	// Validation limits
	authorNameMaxLen = 100
	contentMaxLen    = 5000
)

// Validation errors for comment creation.
var (
	ErrAuthorNameRequired = errors.New("author name is required")
	ErrAuthorNameTooLong  = errors.New("author name must be at most 100 characters")
	ErrContentRequired    = errors.New("content is required")
	ErrContentTooLong     = errors.New("content must be at most 5000 characters")
)

// Comment represents a stored comment on a blog post.
type Comment struct {
	ID         int64     `json:"id"`
	Slug       string    `json:"slug"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	IPHash     string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateCommentInput holds the input data for creating a new comment.
type CreateCommentInput struct {
	Slug    string
	Author  string
	Content string
	IP      string
}

// CommentService defines the interface for comment operations.
type CommentService interface {
	// CreateComment creates a new comment for a post.
	CreateComment(ctx context.Context, input CreateCommentInput) (*Comment, error)
	// GetComments returns all comments for a post in chronological order.
	GetComments(ctx context.Context, slug string) ([]*Comment, error)
	// GetCommentCount returns the comment count for a post.
	GetCommentCount(ctx context.Context, slug string) (int64, error)
}

// commentService implements CommentService.
type commentService struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewCommentService creates a new CommentService instance.
func NewCommentService(db *pgxpool.Pool, redisClient *redis.Client) CommentService {
	return &commentService{
		db:    db,
		redis: redisClient,
	}
}

// CreateComment validates input, inserts a comment into PostgreSQL,
// increments the comment count in Redis cache, and updates post_engagement.
func (s *commentService) CreateComment(ctx context.Context, input CreateCommentInput) (*Comment, error) {
	// Validate author_name (1-100 characters)
	authorLen := utf8.RuneCountInString(input.Author)
	if authorLen == 0 {
		return nil, ErrAuthorNameRequired
	}
	if authorLen > authorNameMaxLen {
		return nil, ErrAuthorNameTooLong
	}

	// Validate content (1-5000 characters)
	contentLen := utf8.RuneCountInString(input.Content)
	if contentLen == 0 {
		return nil, ErrContentRequired
	}
	if contentLen > contentMaxLen {
		return nil, ErrContentTooLong
	}

	// Hash IP using SHA-256 for privacy (reuses hashIP from viewcount.go)
	ipHash := hashIP(input.IP)

	// INSERT into comments table (direct write to PostgreSQL)
	var comment Comment
	err := s.db.QueryRow(ctx, `
		INSERT INTO comments (slug, author_name, content, ip_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, slug, author_name, content, ip_hash, created_at
	`, input.Slug, input.Author, input.Content, ipHash).Scan(
		&comment.ID,
		&comment.Slug,
		&comment.AuthorName,
		&comment.Content,
		&comment.IPHash,
		&comment.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert comment: %w", err)
	}

	// INCR comment count in Redis cache
	if s.redis != nil {
		countKey := fmt.Sprintf("%s%s", commentCountKeyPrefix, input.Slug)
		err = s.redis.Incr(ctx, countKey).Err()
		if err != nil {
			log.Printf("WARNING: Redis error incrementing comment count for %s: %v", input.Slug, err)
			// Non-fatal: comment was already persisted to PostgreSQL
		}
	}

	// Update comment_count in post_engagement table
	_, err = s.db.Exec(ctx, `
		INSERT INTO post_engagement (slug, comment_count, created_at, updated_at)
		VALUES ($1, 1, NOW(), NOW())
		ON CONFLICT (slug)
		DO UPDATE SET comment_count = post_engagement.comment_count + 1, updated_at = NOW()
	`, input.Slug)
	if err != nil {
		log.Printf("WARNING: Failed to update post_engagement comment_count for %s: %v", input.Slug, err)
		// Non-fatal: comment was already persisted
	}

	return &comment, nil
}

// GetComments returns all comments for a post in chronological order (oldest first).
func (s *commentService) GetComments(ctx context.Context, slug string) ([]*Comment, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not available")
	}

	rows, err := s.db.Query(ctx,
		"SELECT id, slug, author_name, content, ip_hash, created_at FROM comments WHERE slug = $1 ORDER BY created_at ASC",
		slug,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		c := &Comment{}
		if err := rows.Scan(&c.ID, &c.Slug, &c.AuthorName, &c.Content, &c.IPHash, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comment rows: %w", err)
	}

	return comments, nil
}

// GetCommentCount returns the comment count for a post.
// It checks the Redis cache first, then falls back to counting in PostgreSQL.
func (s *commentService) GetCommentCount(ctx context.Context, slug string) (int64, error) {
	countKey := fmt.Sprintf("%s%s", commentCountKeyPrefix, slug)

	// Try to get from Redis cache first
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, countKey).Int64()
		if err == nil {
			return cached, nil
		}
		// Cache miss or error - fall through to DB
		if err != redis.Nil {
			log.Printf("WARNING: Redis error getting cached comment count: %v", err)
		}
	}

	// Fallback to PostgreSQL COUNT(*)
	if s.db == nil {
		return 0, fmt.Errorf("database is not available")
	}

	var count int64
	err := s.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM comments WHERE slug = $1", slug,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}

	// Cache the result in Redis
	if s.redis != nil {
		_ = s.redis.Set(ctx, countKey, count, commentCountTTL).Err()
	}

	return count, nil
}
