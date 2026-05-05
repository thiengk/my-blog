package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewCacheMiddleware(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)
	if cm == nil {
		t.Fatal("expected non-nil CacheMiddleware")
	}
	if cm.defaultTTL != 5*time.Minute {
		t.Errorf("expected defaultTTL 5m, got %v", cm.defaultTTL)
	}
}

func TestCacheMiddleware_SkipsNonGET(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)

	router := gin.New()
	router.POST("/api/views/:slug", cm.Middleware(0), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "created"})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/views/test-slug", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCacheMiddleware_SkipsCacheControlNoCache(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)

	handlerCalled := 0
	router := gin.New()
	router.GET("/api/views/:slug", cm.Middleware(0), func(c *gin.Context) {
		handlerCalled++
		c.JSON(http.StatusOK, gin.H{"count": 42})
	})

	// First request with no-cache header
	req := httptest.NewRequest(http.MethodGet, "/api/views/test-slug", nil)
	req.Header.Set("Cache-Control", "no-cache")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if handlerCalled != 1 {
		t.Errorf("expected handler called once, got %d", handlerCalled)
	}
}

func TestCacheMiddleware_FailOpenWhenRedisNil(t *testing.T) {
	// Redis client is nil - should proceed without caching (fail-open)
	cm := NewCacheMiddleware(nil, 5*time.Minute)

	handlerCalled := 0
	router := gin.New()
	router.GET("/api/views/:slug", cm.Middleware(0), func(c *gin.Context) {
		handlerCalled++
		c.JSON(http.StatusOK, gin.H{"count": 10})
	})

	// Multiple requests should all hit the handler (no caching)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/views/test-slug", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i, w.Code)
		}
	}

	if handlerCalled != 3 {
		t.Errorf("expected handler called 3 times (fail-open), got %d", handlerCalled)
	}
}

func TestBuildCacheKey_NoQueryParams(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)

	router := gin.New()
	var capturedKey string
	router.GET("/api/views/:slug", func(c *gin.Context) {
		capturedKey = cm.buildCacheKey(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/views/hello-world", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := "cache:/api/views/hello-world"
	if capturedKey != expected {
		t.Errorf("expected key %q, got %q", expected, capturedKey)
	}
}

func TestBuildCacheKey_WithSortedQueryParams(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)

	router := gin.New()
	var capturedKey string
	router.GET("/api/views", func(c *gin.Context) {
		capturedKey = cm.buildCacheKey(c)
		c.Status(http.StatusOK)
	})

	// Query params in unsorted order
	req := httptest.NewRequest(http.MethodGet, "/api/views?slugs=c,b,a&page=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := "cache:/api/views?page=1&slugs=c,b,a"
	if capturedKey != expected {
		t.Errorf("expected key %q, got %q", expected, capturedKey)
	}
}

func TestBuildCacheKey_DeterministicRegardlessOfParamOrder(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)

	router := gin.New()
	var keys []string
	router.GET("/api/views", func(c *gin.Context) {
		keys = append(keys, cm.buildCacheKey(c))
		c.Status(http.StatusOK)
	})

	// Same params in different order should produce same key
	urls := []string{
		"/api/views?b=2&a=1",
		"/api/views?a=1&b=2",
	}

	for _, url := range urls {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != keys[1] {
		t.Errorf("expected same cache key for different param order: %q vs %q", keys[0], keys[1])
	}
}

func TestInvalidateCache_NilClient(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)
	// Should not panic with nil client
	cm.InvalidateCache("cache:views:*")
}

func TestInvalidateCacheKeys_NilClient(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)
	// Should not panic with nil client
	cm.InvalidateCacheKeys("cache:views:hello-world")
}

func TestInvalidateCacheKeys_EmptyKeys(t *testing.T) {
	cm := NewCacheMiddleware(nil, 5*time.Minute)
	// Should not panic with empty keys
	cm.InvalidateCacheKeys()
}
