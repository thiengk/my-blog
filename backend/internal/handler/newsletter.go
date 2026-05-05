package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// NewsletterHandler handles HTTP requests for newsletter operations.
type NewsletterHandler struct {
	service service.NewsletterService
}

// NewNewsletterHandler creates a new NewsletterHandler with the given service.
func NewNewsletterHandler(svc service.NewsletterService) *NewsletterHandler {
	return &NewsletterHandler{
		service: svc,
	}
}

// RegisterRoutes registers newsletter routes on the given router group.
func (h *NewsletterHandler) RegisterRoutes(rg *gin.RouterGroup) {
	newsletter := rg.Group("/newsletter")
	{
		newsletter.POST("/subscribe", h.Subscribe)
		newsletter.POST("/unsubscribe", h.Unsubscribe)
		newsletter.GET("/verify/:token", h.VerifyEmail)
	}
}

// subscribeRequest represents the request body for subscribing.
type subscribeRequest struct {
	Email string `json:"email" binding:"required"`
}

// unsubscribeRequest represents the request body for unsubscribing.
type unsubscribeRequest struct {
	Email string `json:"email" binding:"required"`
}

// Subscribe handles POST /api/newsletter/subscribe
// Registers a new email for the newsletter.
func (h *NewsletterHandler) Subscribe(c *gin.Context) {
	var req subscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "email is required",
		})
		return
	}

	err := h.service.Subscribe(c.Request.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "invalid email address",
			})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already subscribed",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to subscribe",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription successful, please check your email to verify",
	})
}

// Unsubscribe handles POST /api/newsletter/unsubscribe
// Marks a subscriber as unsubscribed.
func (h *NewsletterHandler) Unsubscribe(c *gin.Context) {
	var req unsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": "email is required",
		})
		return
	}

	err := h.service.Unsubscribe(c.Request.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "invalid email address",
			})
		case errors.Is(err, service.ErrSubscriberNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscriber not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to unsubscribe",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully unsubscribed",
	})
}

// VerifyEmail handles GET /api/newsletter/verify/:token
// Verifies a subscriber's email using the verification token.
func (h *NewsletterHandler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "verification token is required",
		})
		return
	}

	err := h.service.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "invalid or expired verification token",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to verify email",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "email verified successfully",
	})
}
