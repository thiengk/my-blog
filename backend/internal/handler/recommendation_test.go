package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// --- Mock RecommendationService ---

type mockRecommendationService struct {
	getTopPostsFunc          func(ctx context.Context, limit int) ([]*service.RankedPost, error)
	recalculateRankingsFunc  func(ctx context.Context) error
}

func (m *mockRecommendationService) GetTopPosts(ctx context.Context, limit int) ([]*service.RankedPost, error) {
	if m.getTopPostsFunc != nil {
		return m.getTopPostsFunc(ctx, limit)
	}
	return []*service.RankedPost{}, nil
}

func (m *mockRecommendationService) RecalculateRankings(ctx context.Context) error {
	if m.recalculateRankingsFunc != nil {
		return m.recalculateRankingsFunc(ctx)
	}
	return nil
}

// --- RecommendationHandler Tests ---

func TestRecommendationHandler_GetRecommendations_DefaultLimit(t *testing.T) {
	mock := &mockRecommendationService{
		getTopPostsFunc: func(ctx context.Context, limit int) ([]*service.RankedPost, error) {
			if limit != 10 {
				t.Errorf("expected default limit 10, got %d", limit)
			}
			return []*service.RankedPost{
				{Slug: "popular-post", EngagementScore: 25.0, Likes: 10, Comments: 5, Shares: 2},
				{Slug: "another-post", EngagementScore: 15.0, Likes: 5, Comments: 3, Shares: 1},
			}, nil
		},
	}

	h := NewRecommendationHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	posts, ok := resp["posts"].([]interface{})
	if !ok {
		t.Fatal("expected 'posts' to be an array")
	}

	if len(posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(posts))
	}

	first := posts[0].(map[string]interface{})
	if first["slug"] != "popular-post" {
		t.Errorf("expected first post slug 'popular-post', got '%v'", first["slug"])
	}
	if first["engagement_score"].(float64) != 25.0 {
		t.Errorf("expected engagement_score 25.0, got %v", first["engagement_score"])
	}
}

func TestRecommendationHandler_GetRecommendations_CustomLimit(t *testing.T) {
	mock := &mockRecommendationService{
		getTopPostsFunc: func(ctx context.Context, limit int) ([]*service.RankedPost, error) {
			if limit != 5 {
				t.Errorf("expected limit 5, got %d", limit)
			}
			return []*service.RankedPost{
				{Slug: "post-1", EngagementScore: 30.0, Likes: 15, Comments: 5, Shares: 2},
			}, nil
		},
	}

	h := NewRecommendationHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	posts, ok := resp["posts"].([]interface{})
	if !ok {
		t.Fatal("expected 'posts' to be an array")
	}

	if len(posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(posts))
	}
}

func TestRecommendationHandler_GetRecommendations_LimitClamped(t *testing.T) {
	mock := &mockRecommendationService{
		getTopPostsFunc: func(ctx context.Context, limit int) ([]*service.RankedPost, error) {
			// Handler clamps limit to 50 before calling service
			if limit != 50 {
				t.Errorf("expected clamped limit 50, got %d", limit)
			}
			return []*service.RankedPost{}, nil
		},
	}

	h := NewRecommendationHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=100", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestRecommendationHandler_GetRecommendations_InvalidLimit(t *testing.T) {
	mock := &mockRecommendationService{}

	h := NewRecommendationHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "limit must be a positive integer" {
		t.Errorf("expected error 'limit must be a positive integer', got '%v'", resp["error"])
	}
}

func TestRecommendationHandler_GetRecommendations_NegativeLimit(t *testing.T) {
	mock := &mockRecommendationService{}

	h := NewRecommendationHandler(mock)
	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/recommendations?limit=-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
