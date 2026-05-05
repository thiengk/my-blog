package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHealthHandler_Check_AllNil(t *testing.T) {
	// When both db and redis are nil, status should be "unhealthy"
	h := NewHealthHandler(nil, nil)

	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Status != "unhealthy" {
		t.Errorf("expected overall status 'unhealthy', got '%s'", resp.Status)
	}

	if resp.Services["postgres"].Status != "down" {
		t.Errorf("expected postgres status 'down', got '%s'", resp.Services["postgres"].Status)
	}

	if resp.Services["redis"].Status != "down" {
		t.Errorf("expected redis status 'down', got '%s'", resp.Services["redis"].Status)
	}

	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestDetermineOverallStatus(t *testing.T) {
	tests := []struct {
		name     string
		services map[string]ServiceStatus
		expected string
	}{
		{
			name: "all services up returns healthy",
			services: map[string]ServiceStatus{
				"postgres": {Status: "up", LatencyMs: 5},
				"redis":    {Status: "up", LatencyMs: 2},
			},
			expected: "healthy",
		},
		{
			name: "all services down returns unhealthy",
			services: map[string]ServiceStatus{
				"postgres": {Status: "down", LatencyMs: 0},
				"redis":    {Status: "down", LatencyMs: 0},
			},
			expected: "unhealthy",
		},
		{
			name: "some services down returns degraded",
			services: map[string]ServiceStatus{
				"postgres": {Status: "up", LatencyMs: 5},
				"redis":    {Status: "down", LatencyMs: 0},
			},
			expected: "degraded",
		},
		{
			name: "postgres down redis up returns degraded",
			services: map[string]ServiceStatus{
				"postgres": {Status: "down", LatencyMs: 0},
				"redis":    {Status: "up", LatencyMs: 2},
			},
			expected: "degraded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineOverallStatus(tt.services)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	// Verify the JSON response has the expected structure
	h := NewHealthHandler(nil, nil)

	router := gin.New()
	api := router.Group("/api")
	h.RegisterRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Check top-level fields exist
	if _, ok := raw["status"]; !ok {
		t.Error("response missing 'status' field")
	}
	if _, ok := raw["services"]; !ok {
		t.Error("response missing 'services' field")
	}
	if _, ok := raw["timestamp"]; !ok {
		t.Error("response missing 'timestamp' field")
	}

	// Check services structure
	services, ok := raw["services"].(map[string]interface{})
	if !ok {
		t.Fatal("'services' is not an object")
	}

	for _, svcName := range []string{"postgres", "redis"} {
		svc, ok := services[svcName].(map[string]interface{})
		if !ok {
			t.Errorf("service '%s' is not an object", svcName)
			continue
		}
		if _, ok := svc["status"]; !ok {
			t.Errorf("service '%s' missing 'status' field", svcName)
		}
		if _, ok := svc["latency_ms"]; !ok {
			t.Errorf("service '%s' missing 'latency_ms' field", svcName)
		}
	}
}
