package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// --- Mock CommentService ---

type mockCommentService struct {
	createCommentFunc   func(ctx context.Context, input service.CreateCommentInput) (*service.Comment, error)
	getCommentsFunc     func(ctx context.Context, slug string) ([]*service.Comment, error)
	getCommentCountFunc func(ctx context.Context, slug string) (int64, error)
}

func (m *mockCommentService) CreateComment(ctx context.Context, input service.CreateCommentInput) (*service.Comment, error) {
	if m.createCommentFunc != nil {
		return m.createCommentFunc(ctx, input)
	}
	return &service.Comment{}, nil
}

func (m *mockCommentService) GetComments(ctx context.Context, slug string) ([]*service.Comment, error) {
	if m.getCommentsFunc != nil {
		return m.getCommentsFunc(ctx, slug)
	}
	return []*service.Comment{}, nil
}

func (m *mockCommentService) GetCommentCount(ctx context.Context, slug string) (int64, error) {
	if m.getCommentCountFunc != nil {
		return m.getCommentCountFunc(ctx, slug)
	}
	return 0, nil
}

// --- CommentHandler Tests ---

func TestCommentHandler_CreateComment_Success(t *testing.T) {
	now := time.Now()
	mock := &mockCommentService{
		createCommentFunc: func(ctx context.Context, input service.CreateCommentInput) (*service.Comment, error) {
			if input.Slug != "test-post" {
				t.Errorf("expected slug 'test-post', got '%s'", input.Slug)
			}
			if input.Author != "John Doe" {
				t.Errorf("expected author 'John Doe', got '%s'", input.Author)
			}
			if input.Content != "Great article!" {
				t.Errorf("expected content 'Great article!', got '%s'", input.Content)
			}
			return &service.Comment{
				ID:         1,
				Slug:       input.Slug,
				AuthorName: input.Author,
				Content:    input.Content,
				CreatedAt:  now,
			}, nil
		},
	}

	h := NewCommentHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	body := `{"author_name":"John Doe","content":"Great article!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments/test-post", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["author_name"] != "John Doe" {
		t.Errorf("expected author_name 'John Doe', got '%v'", resp["author_name"])
	}
	if resp["content"] != "Great article!" {
		t.Errorf("expected content 'Great article!', got '%v'", resp["content"])
	}
	if resp["slug"] != "test-post" {
		t.Errorf("expected slug 'test-post', got '%v'", resp["slug"])
	}
}

func TestCommentHandler_CreateComment_ValidationError_EmptyAuthor(t *testing.T) {
	mock := &mockCommentService{
		createCommentFunc: func(ctx context.Context, input service.CreateCommentInput) (*service.Comment, error) {
			return nil, service.ErrAuthorNameRequired
		},
	}

	h := NewCommentHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	body := `{"author_name":"","content":"Some content"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments/test-post", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "author name is required" {
		t.Errorf("expected error 'author name is required', got '%v'", resp["error"])
	}
}

func TestCommentHandler_CreateComment_ValidationError_ContentTooLong(t *testing.T) {
	mock := &mockCommentService{
		createCommentFunc: func(ctx context.Context, input service.CreateCommentInput) (*service.Comment, error) {
			return nil, service.ErrContentTooLong
		},
	}

	h := NewCommentHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	body := `{"author_name":"John","content":"too long content"}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments/test-post", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "content must be at most 5000 characters" {
		t.Errorf("expected error 'content must be at most 5000 characters', got '%v'", resp["error"])
	}
}

func TestCommentHandler_CreateComment_InvalidJSON(t *testing.T) {
	mock := &mockCommentService{}

	h := NewCommentHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/comments/test-post", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCommentHandler_GetComments_Success(t *testing.T) {
	now := time.Now()
	mock := &mockCommentService{
		getCommentsFunc: func(ctx context.Context, slug string) ([]*service.Comment, error) {
			if slug != "test-post" {
				t.Errorf("expected slug 'test-post', got '%s'", slug)
			}
			return []*service.Comment{
				{
					ID:         1,
					Slug:       "test-post",
					AuthorName: "Alice",
					Content:    "First comment",
					CreatedAt:  now.Add(-time.Hour),
				},
				{
					ID:         2,
					Slug:       "test-post",
					AuthorName: "Bob",
					Content:    "Second comment",
					CreatedAt:  now,
				},
			}, nil
		},
	}

	h := NewCommentHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/comments/test-post", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	comments, ok := resp["comments"].([]interface{})
	if !ok {
		t.Fatal("expected 'comments' to be an array")
	}

	if len(comments) != 2 {
		t.Errorf("expected 2 comments, got %d", len(comments))
	}

	first := comments[0].(map[string]interface{})
	if first["author_name"] != "Alice" {
		t.Errorf("expected first comment author 'Alice', got '%v'", first["author_name"])
	}

	second := comments[1].(map[string]interface{})
	if second["author_name"] != "Bob" {
		t.Errorf("expected second comment author 'Bob', got '%v'", second["author_name"])
	}
}

func TestCommentHandler_GetComments_Empty(t *testing.T) {
	mock := &mockCommentService{
		getCommentsFunc: func(ctx context.Context, slug string) ([]*service.Comment, error) {
			return nil, nil
		},
	}

	h := NewCommentHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/comments/empty-post", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	comments, ok := resp["comments"].([]interface{})
	if !ok {
		t.Fatal("expected 'comments' to be an array")
	}

	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}
