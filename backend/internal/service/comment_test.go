package service

import (
	"context"
	"strings"
	"testing"
)

// TestCreateComment_ValidationErrors tests that CreateComment returns appropriate
// validation errors for invalid inputs. These tests exercise the validation logic
// directly without needing a database connection.
func TestCreateComment_ValidationErrors(t *testing.T) {
	// Create service with nil DB and Redis — validation happens before DB access
	svc := &commentService{db: nil, redis: nil}
	ctx := context.Background()

	tests := []struct {
		name        string
		input       CreateCommentInput
		expectedErr error
	}{
		{
			name: "empty author name returns ErrAuthorNameRequired",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  "",
				Content: "This is a valid comment.",
				IP:      "192.168.1.1",
			},
			expectedErr: ErrAuthorNameRequired,
		},
		{
			name: "whitespace-only author is treated as non-empty (passes author check)",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  "   ",
				Content: "Valid content",
				IP:      "192.168.1.1",
			},
			// Whitespace-only author has length > 0, so it passes validation.
			// It will fail at DB insert (nil db), but that's expected.
			expectedErr: nil,
		},
		{
			name: "author name too long returns ErrAuthorNameTooLong",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  strings.Repeat("a", 101),
				Content: "Valid content",
				IP:      "192.168.1.1",
			},
			expectedErr: ErrAuthorNameTooLong,
		},
		{
			name: "author name exactly 100 chars passes validation",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  strings.Repeat("b", 100),
				Content: "Valid content",
				IP:      "192.168.1.1",
			},
			// Passes validation, fails at DB insert (nil db)
			expectedErr: nil,
		},
		{
			name: "empty content returns ErrContentRequired",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  "John",
				Content: "",
				IP:      "192.168.1.1",
			},
			expectedErr: ErrContentRequired,
		},
		{
			name: "content too long returns ErrContentTooLong",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  "John",
				Content: strings.Repeat("x", 5001),
				IP:      "192.168.1.1",
			},
			expectedErr: ErrContentTooLong,
		},
		{
			name: "content exactly 5000 chars passes validation",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  "John",
				Content: strings.Repeat("y", 5000),
				IP:      "192.168.1.1",
			},
			// Passes validation, fails at DB insert (nil db)
			expectedErr: nil,
		},
		{
			name: "unicode author name within limit passes validation",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  strings.Repeat("日", 100), // 100 runes, each multi-byte
				Content: "Valid content",
				IP:      "192.168.1.1",
			},
			// Passes validation (100 runes = exactly at limit)
			expectedErr: nil,
		},
		{
			name: "unicode author name exceeding limit returns ErrAuthorNameTooLong",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  strings.Repeat("日", 101), // 101 runes
				Content: "Valid content",
				IP:      "192.168.1.1",
			},
			expectedErr: ErrAuthorNameTooLong,
		},
		{
			name: "unicode content exceeding limit returns ErrContentTooLong",
			input: CreateCommentInput{
				Slug:    "test-post",
				Author:  "John",
				Content: strings.Repeat("中", 5001), // 5001 runes
				IP:      "192.168.1.1",
			},
			expectedErr: ErrContentTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateComment(ctx, tt.input)

			if tt.expectedErr != nil {
				// We expect a specific validation error
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedErr)
				}
				if err != tt.expectedErr {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				// We expect validation to pass; the error (if any) should be
				// from the DB layer (nil pointer), not a validation error.
				if err == ErrAuthorNameRequired || err == ErrAuthorNameTooLong ||
					err == ErrContentRequired || err == ErrContentTooLong {
					t.Errorf("unexpected validation error: %v", err)
				}
				// Any other error (e.g., nil DB panic recovery) is acceptable
				// since we're only testing validation logic here.
			}
		})
	}
}

// TestCreateComment_ValidationOrder verifies that author validation happens
// before content validation (empty author + empty content → ErrAuthorNameRequired).
func TestCreateComment_ValidationOrder(t *testing.T) {
	svc := &commentService{db: nil, redis: nil}
	ctx := context.Background()

	_, err := svc.CreateComment(ctx, CreateCommentInput{
		Slug:    "test-post",
		Author:  "",
		Content: "",
		IP:      "192.168.1.1",
	})

	if err != ErrAuthorNameRequired {
		t.Errorf("expected ErrAuthorNameRequired when both fields empty, got %v", err)
	}
}

// TestGetComments_NilDB verifies that GetComments returns an error when DB is nil.
func TestGetComments_NilDB(t *testing.T) {
	svc := &commentService{db: nil, redis: nil}
	ctx := context.Background()

	comments, err := svc.GetComments(ctx, "test-post")
	if err == nil {
		t.Fatal("expected error when DB is nil, got nil")
	}
	if comments != nil {
		t.Errorf("expected nil comments when DB is nil, got %v", comments)
	}
}

// TestGetCommentCount_NilDBAndRedis verifies that GetCommentCount returns an error
// when both Redis (cache miss) and DB are nil.
func TestGetCommentCount_NilDBAndRedis(t *testing.T) {
	svc := &commentService{db: nil, redis: nil}
	ctx := context.Background()

	count, err := svc.GetCommentCount(ctx, "test-post")
	if err == nil {
		t.Fatal("expected error when DB is nil, got nil")
	}
	if count != 0 {
		t.Errorf("expected count 0 on error, got %d", count)
	}
}

// TestCommentConstants verifies that the validation constants are set correctly.
func TestCommentConstants(t *testing.T) {
	if authorNameMaxLen != 100 {
		t.Errorf("expected authorNameMaxLen = 100, got %d", authorNameMaxLen)
	}
	if contentMaxLen != 5000 {
		t.Errorf("expected contentMaxLen = 5000, got %d", contentMaxLen)
	}
}

// TestCommentCountKeyPrefix verifies the Redis key format for comment counts.
func TestCommentCountKeyPrefix(t *testing.T) {
	expected := "comment:count:"
	if commentCountKeyPrefix != expected {
		t.Errorf("expected commentCountKeyPrefix = %q, got %q", expected, commentCountKeyPrefix)
	}
}
