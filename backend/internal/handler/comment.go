package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// CommentHandler handles HTTP requests for comment operations.
type CommentHandler struct {
	service service.CommentService
}

// NewCommentHandler creates a new CommentHandler with the given service.
func NewCommentHandler(svc service.CommentService) *CommentHandler {
	return &CommentHandler{
		service: svc,
	}
}

// RegisterRoutes registers comment routes on the given router group.
func (h *CommentHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/comments/:slug", h.CreateComment)
	rg.GET("/comments/:slug", h.GetComments)
}

// createCommentRequest represents the JSON body for creating a comment.
type createCommentRequest struct {
	AuthorName string `json:"author_name"`
	Content    string `json:"content"`
}

// CreateComment handles POST /api/comments/:slug
// Creates a new comment for the given post slug.
func (h *CommentHandler) CreateComment(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ip := extractClientIPFromRequest(c)

	input := service.CreateCommentInput{
		Slug:    slug,
		Author:  req.AuthorName,
		Content: req.Content,
		IP:      ip,
	}

	comment, err := h.service.CreateComment(c.Request.Context(), input)
	if err != nil {
		// Check for validation errors from the service
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create comment",
		})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// GetComments handles GET /api/comments/:slug
// Returns all comments for the given post slug in chronological order.
func (h *CommentHandler) GetComments(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "slug is required",
		})
		return
	}

	comments, err := h.service.GetComments(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get comments",
		})
		return
	}

	// Return empty array instead of null when no comments exist
	if comments == nil {
		comments = []*service.Comment{}
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": comments,
	})
}

// isValidationError checks if the error is a known validation error from the comment service.
func isValidationError(err error) bool {
	return errors.Is(err, service.ErrAuthorNameRequired) ||
		errors.Is(err, service.ErrAuthorNameTooLong) ||
		errors.Is(err, service.ErrContentRequired) ||
		errors.Is(err, service.ErrContentTooLong)
}
