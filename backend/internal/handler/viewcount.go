package handler

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// ViewCountHandler handles HTTP requests for view count operations.
type ViewCountHandler struct {
	service service.ViewCountService
}

// NewViewCountHandler creates a new ViewCountHandler with the given service.
func NewViewCountHandler(svc service.ViewCountService) *ViewCountHandler {
	return &ViewCountHandler{
		service: svc,
	}
}

// RegisterRoutes registers view count routes on the given router group.
func (h *ViewCountHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/views/:slug", h.IncrementView)
	rg.GET("/views/:slug", h.GetViewCount)
	rg.GET("/views", h.GetBulkViewCounts)
}

// IncrementView handles POST /api/views/:slug
// Records a page view for the given slug using the client's IP.
func (h *ViewCountHandler) IncrementView(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	ip := extractClientIPFromRequest(c)

	counted, err := h.service.RecordView(c.Request.Context(), slug, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record view",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"counted": counted,
		"slug":    slug,
	})
}

// GetViewCount handles GET /api/views/:slug
// Returns the view count for a single post.
func (h *ViewCountHandler) GetViewCount(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	count, err := h.service.GetCount(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get view count",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slug":  slug,
		"count": count,
	})
}

// GetBulkViewCounts handles GET /api/views?slugs=a,b,c
// Returns view counts for multiple posts.
func (h *ViewCountHandler) GetBulkViewCounts(c *gin.Context) {
	slugsParam := c.Query("slugs")
	if slugsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slugs query parameter is required",
		})
		return
	}

	slugs := parseSlugs(slugsParam)
	if len(slugs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one valid slug is required",
		})
		return
	}

	// Limit the number of slugs to prevent abuse
	if len(slugs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "maximum 50 slugs allowed per request",
		})
		return
	}

	counts, err := h.service.GetBulkCounts(c.Request.Context(), slugs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get view counts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"counts": counts,
	})
}

// parseSlugs splits a comma-separated string into a slice of trimmed, non-empty slugs.
func parseSlugs(input string) []string {
	parts := strings.Split(input, ",")
	slugs := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			slugs = append(slugs, trimmed)
		}
	}
	return slugs
}

// extractClientIPFromRequest extracts the real client IP from the request,
// checking X-Forwarded-For, X-Real-IP headers, and falling back to RemoteAddr.
func extractClientIPFromRequest(c *gin.Context) string {
	// Check X-Forwarded-For header (may contain multiple IPs: client, proxy1, proxy2)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (strip port if present)
	remoteAddr := c.Request.RemoteAddr
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}
