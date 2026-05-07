package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// --- Mock EngagementService ---

type mockEngagementService struct {
	recordLikeFunc    func(ctx context.Context, slug string, ip string) (bool, error)
	recordShareFunc   func(ctx context.Context, slug string, ip string, platform string) (bool, error)
	getCountsFunc     func(ctx context.Context, slug string) (*service.EngagementCounts, error)
	getBulkCountsFunc func(ctx context.Context, slugs []string) (map[string]*service.EngagementCounts, error)
	flushBatchFunc    func(ctx context.Context) error
}

func (m *mockEngagementService) RecordLike(ctx context.Context, slug string, ip string) (bool, error) {
	if m.recordLikeFunc != nil {
		return m.recordLikeFunc(ctx, slug, ip)
	}
	return true, nil
}

func (m *mockEngagementService) RecordShare(ctx context.Context, slug string, ip string, platform string) (bool, error) {
	if m.recordShareFunc != nil {
		return m.recordShareFunc(ctx, slug, ip, platform)
	}
	return true, nil
}

func (m *mockEngagementService) GetCounts(ctx context.Context, slug string) (*service.EngagementCounts, error) {
	if m.getCountsFunc != nil {
		return m.getCountsFunc(ctx, slug)
	}
	return &service.EngagementCounts{}, nil
}

func (m *mockEngagementService) GetBulkCounts(ctx context.Context, slugs []string) (map[string]*service.EngagementCounts, error) {
	if m.getBulkCountsFunc != nil {
		return m.getBulkCountsFunc(ctx, slugs)
	}
	return map[string]*service.EngagementCounts{}, nil
}

func (m *mockEngagementService) FlushBatch(ctx context.Context) error {
	if m.flushBatchFunc != nil {
		return m.flushBatchFunc(ctx)
	}
	return nil
}

// --- EngagementHandler Tests ---

func TestEngagementHandler_RecordLike_Success(t *testing.T) {
	mock := &mockEngagementService{
		recordLikeFunc: func(ctx context.Context, slug string, ip string) (bool, error) {
			if slug != "test-post" {
				t.Errorf("expected slug 'test-post', got '%s'", slug)
			}
			return true, nil
		},
	}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodPost, "/api/engagement/like/test-post", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	counted, ok := resp["counted"].(bool)
	if !ok || !counted {
		t.Errorf("expected counted=true, got %v", resp["counted"])
	}
}

func TestEngagementHandler_RecordLike_Duplicate(t *testing.T) {
	mock := &mockEngagementService{
		recordLikeFunc: func(ctx context.Context, slug string, ip string) (bool, error) {
			// Simulate duplicate like - not counted
			return false, nil
		},
	}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodPost, "/api/engagement/like/test-post", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	counted, ok := resp["counted"].(bool)
	if !ok || counted {
		t.Errorf("expected counted=false, got %v", resp["counted"])
	}
}

func TestEngagementHandler_RecordShare_Success(t *testing.T) {
	mock := &mockEngagementService{
		recordShareFunc: func(ctx context.Context, slug string, ip string, platform string) (bool, error) {
			if slug != "test-post" {
				t.Errorf("expected slug 'test-post', got '%s'", slug)
			}
			if platform != "twitter" {
				t.Errorf("expected platform 'twitter', got '%s'", platform)
			}
			return true, nil
		},
	}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	body := `{"platform":"twitter"}`
	req := httptest.NewRequest(http.MethodPost, "/api/engagement/share/test-post", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	counted, ok := resp["counted"].(bool)
	if !ok || !counted {
		t.Errorf("expected counted=true, got %v", resp["counted"])
	}
}

func TestEngagementHandler_RecordShare_MissingPlatform(t *testing.T) {
	mock := &mockEngagementService{}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/engagement/share/test-post", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "platform is required" {
		t.Errorf("expected error 'platform is required', got '%v'", resp["error"])
	}
}

func TestEngagementHandler_GetCounts(t *testing.T) {
	mock := &mockEngagementService{
		getCountsFunc: func(ctx context.Context, slug string) (*service.EngagementCounts, error) {
			if slug != "my-post" {
				t.Errorf("expected slug 'my-post', got '%s'", slug)
			}
			return &service.EngagementCounts{
				Likes:    10,
				Comments: 5,
				Shares:   3,
			}, nil
		},
	}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/engagement/my-post", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["slug"] != "my-post" {
		t.Errorf("expected slug 'my-post', got '%v'", resp["slug"])
	}
	if resp["likes"].(float64) != 10 {
		t.Errorf("expected likes=10, got %v", resp["likes"])
	}
	if resp["comments"].(float64) != 5 {
		t.Errorf("expected comments=5, got %v", resp["comments"])
	}
	if resp["shares"].(float64) != 3 {
		t.Errorf("expected shares=3, got %v", resp["shares"])
	}
}

func TestEngagementHandler_GetBulkCounts(t *testing.T) {
	mock := &mockEngagementService{
		getBulkCountsFunc: func(ctx context.Context, slugs []string) (map[string]*service.EngagementCounts, error) {
			if len(slugs) != 2 {
				t.Errorf("expected 2 slugs, got %d", len(slugs))
			}
			return map[string]*service.EngagementCounts{
				"post-1": {Likes: 5, Comments: 2, Shares: 1},
				"post-2": {Likes: 8, Comments: 3, Shares: 4},
			}, nil
		},
	}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/engagement?slugs=post-1,post-2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	counts, ok := resp["counts"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'counts' to be an object")
	}

	if len(counts) != 2 {
		t.Errorf("expected 2 entries in counts, got %d", len(counts))
	}
}

func TestEngagementHandler_GetBulkCounts_MissingSlugs(t *testing.T) {
	mock := &mockEngagementService{}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/engagement?slugs=", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestEngagementHandler_GetBulkCounts_NoSlugsParam(t *testing.T) {
	mock := &mockEngagementService{}

	h := NewEngagementHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/engagement", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
