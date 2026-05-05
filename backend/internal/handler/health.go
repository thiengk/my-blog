package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// ServiceStatus represents the health status of an individual service.
type ServiceStatus struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
}

// HealthResponse represents the overall health check response.
type HealthResponse struct {
	Status    string                   `json:"status"`
	Services  map[string]ServiceStatus `json:"services"`
	Timestamp string                   `json:"timestamp"`
}

// HealthHandler handles HTTP requests for the health check endpoint.
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewHealthHandler creates a new HealthHandler with the given database pool and Redis client.
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// RegisterRoutes registers health check routes on the given router group.
func (h *HealthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", h.Check)
}

// Check handles GET /api/health
// It pings PostgreSQL and Redis, returning the status of each service.
func (h *HealthHandler) Check(c *gin.Context) {
	services := make(map[string]ServiceStatus)

	// Check PostgreSQL
	pgStatus := h.checkPostgres(c.Request.Context())
	services["postgres"] = pgStatus

	// Check Redis
	redisStatus := h.checkRedis(c.Request.Context())
	services["redis"] = redisStatus

	// Determine overall status
	overallStatus := determineOverallStatus(services)

	resp := HealthResponse{
		Status:    overallStatus,
		Services:  services,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, resp)
}

// checkPostgres pings the PostgreSQL database and returns its status.
func (h *HealthHandler) checkPostgres(ctx context.Context) ServiceStatus {
	if h.db == nil {
		return ServiceStatus{
			Status:    "down",
			LatencyMs: 0,
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()
	err := h.db.Ping(pingCtx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return ServiceStatus{
			Status:    "down",
			LatencyMs: latency,
		}
	}

	return ServiceStatus{
		Status:    "up",
		LatencyMs: latency,
	}
}

// checkRedis pings the Redis client and returns its status.
func (h *HealthHandler) checkRedis(ctx context.Context) ServiceStatus {
	if h.redis == nil {
		return ServiceStatus{
			Status:    "down",
			LatencyMs: 0,
		}
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	start := time.Now()
	err := h.redis.Ping(pingCtx).Err()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return ServiceStatus{
			Status:    "down",
			LatencyMs: latency,
		}
	}

	return ServiceStatus{
		Status:    "up",
		LatencyMs: latency,
	}
}

// determineOverallStatus returns the overall health status based on individual service statuses.
// "healthy" = all services up, "degraded" = some services down, "unhealthy" = all services down.
func determineOverallStatus(services map[string]ServiceStatus) string {
	upCount := 0
	totalCount := len(services)

	for _, svc := range services {
		if svc.Status == "up" {
			upCount++
		}
	}

	switch {
	case upCount == totalCount:
		return "healthy"
	case upCount == 0:
		return "unhealthy"
	default:
		return "degraded"
	}
}
