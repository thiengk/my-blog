package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// ShareRequest represents the request body for recording a share.
type ShareRequest struct {
	Platform string `json:"platform" binding:"required"`
}

// EngagementHandler handles HTTP requests for engagement operations (likes, shares, counts).
type EngagementHandler struct {
	service service.EngagementService
}

// NewEngagementHandler creates a new EngagementHandler with the given service.
func NewEngagementHandler(svc service.EngagementService) *EngagementHandler {
	return &EngagementHandler{
		service: svc,
	}
}

// RegisterRoutes registers engagement routes on the given router group.
func (h *EngagementHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/engagement/like/:slug", h.RecordLike)
	rg.POST("/engagement/share/:slug", h.RecordShare)
	rg.GET("/engagement/:slug", h.GetCounts)
	rg.GET("/engagement", h.GetBulkCounts)
}

// RecordLike handles POST /api/engagement/like/:slug
// Records a like for the given slug using the client's IP for deduplication.
func (h *EngagementHandler) RecordLike(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	ip := extractClientIPFromRequest(c)

	counted, err := h.service.RecordLike(c.Request.Context(), slug, ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record like",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"counted": counted,
	})
}

// RecordShare handles POST /api/engagement/share/:slug
// Records a share for the given slug with platform info from the request body.
func (h *EngagementHandler) RecordShare(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	var req ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "platform is required",
		})
		return
	}

	ip := extractClientIPFromRequest(c)

	counted, err := h.service.RecordShare(c.Request.Context(), slug, ip, req.Platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record share",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"counted": counted,
	})
}

// GetCounts handles GET /api/engagement/:slug
// Returns engagement counts (likes, comments, shares) for a single post.
func (h *EngagementHandler) GetCounts(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	counts, err := h.service.GetCounts(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get engagement counts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"slug":     slug,
		"likes":    counts.Likes,
		"comments": counts.Comments,
		"shares":   counts.Shares,
	})
}

// GetBulkCounts handles GET /api/engagement?slugs=a,b,c
// Returns engagement counts for multiple posts. Maximum 50 slugs per request.
func (h *EngagementHandler) GetBulkCounts(c *gin.Context) {
	slugsParam := c.Query("slugs")
	if slugsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slugs query parameter is required",
		})
		return
	}

	slugs := parseEngagementSlugs(slugsParam)
	if len(slugs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "at least one valid slug is required",
		})
		return
	}

	// Validate maximum 50 slugs per request
	if len(slugs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "maximum 50 slugs allowed per request",
		})
		return
	}

	counts, err := h.service.GetBulkCounts(c.Request.Context(), slugs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get engagement counts",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"counts": counts,
	})
}

// parseEngagementSlugs splits a comma-separated string into a slice of trimmed, non-empty slugs.
func parseEngagementSlugs(input string) []string {
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
