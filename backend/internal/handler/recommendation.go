package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// RecommendationHandler handles HTTP requests for post recommendations.
type RecommendationHandler struct {
	service service.RecommendationService
}

// NewRecommendationHandler creates a new RecommendationHandler with the given service.
func NewRecommendationHandler(svc service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{
		service: svc,
	}
}

// RegisterRoutes registers recommendation routes on the given router group.
func (h *RecommendationHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/recommendations", h.GetRecommendations)
}

// GetRecommendations handles GET /api/recommendations?limit=10
// Returns the top recommended posts sorted by engagement score.
func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	limit := 10 // default

	if limitParam := c.Query("limit"); limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}
		limit = parsed
	}

	// Clamp limit to maximum of 50
	if limit > 50 {
		limit = 50
	}

	posts, err := h.service.GetTopPosts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get recommendations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
	})
}
