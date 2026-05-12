package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/service"
)

// CreateMemberRequest represents the request body for creating a member.
type CreateMemberRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateMemberRequest represents the request body for updating a member.
type UpdateMemberRequest struct {
	Name     string `json:"name" binding:"required"`
	IsActive bool   `json:"is_active"`
}

// UpdateParticipationsRequest represents the request body for updating participations.
type UpdateParticipationsRequest struct {
	MemberID  int64 `json:"member_id" binding:"required"`
	Breakfast bool  `json:"breakfast"`
	Lunch     bool  `json:"lunch"`
}

// RecordPaymentRequest represents the request body for recording a payment.
type RecordPaymentRequest struct {
	MemberID int64  `json:"member_id" binding:"required"`
	MealType string `json:"meal_type" binding:"required"`
	Date     string `json:"date" binding:"required"`
}

// MealHandler handles HTTP requests for meal payment scheduling.
type MealHandler struct {
	service service.MealService
}

// NewMealHandler creates a new MealHandler with the given service.
func NewMealHandler(svc service.MealService) *MealHandler {
	return &MealHandler{service: svc}
}

// RegisterRoutes registers meal routes on the given router group.
// The router group should already have auth middleware applied.
func (h *MealHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/meals/members", h.GetMembers)
	rg.POST("/meals/members", h.CreateMember)
	rg.PUT("/meals/members/:id", h.UpdateMember)
	rg.DELETE("/meals/members/:id", h.DeleteMember)
	rg.GET("/meals/participations", h.GetParticipations)
	rg.PUT("/meals/participations", h.UpdateParticipations)
	rg.GET("/meals/next-payer", h.GetNextPayer)
	rg.GET("/meals/payments", h.GetPayments)
	rg.POST("/meals/payments", h.RecordPayment)
	rg.DELETE("/meals/payments/:id", h.UndoPayment)
	rg.GET("/meals/stats", h.GetStats)
}

// GetMembers handles GET /api/meals/members
func (h *MealHandler) GetMembers(c *gin.Context) {
	members, err := h.service.GetMembers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

// CreateMember handles POST /api/meals/members
func (h *MealHandler) CreateMember(c *gin.Context) {
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	member, err := h.service.CreateMember(c.Request.Context(), req.Name)
	if err != nil {
		switch err.Error() {
		case "member already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "member already exists"})
		case "name is required":
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"member": member})
}

// UpdateMember handles PUT /api/meals/members/:id
func (h *MealHandler) UpdateMember(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	member, err := h.service.UpdateMember(c.Request.Context(), id, req.Name, req.IsActive)
	if err != nil {
		switch err.Error() {
		case "member not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case "member already exists":
			c.JSON(http.StatusConflict, gin.H{"error": "member already exists"})
		case "name is required":
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"member": member})
}

// DeleteMember handles DELETE /api/meals/members/:id
func (h *MealHandler) DeleteMember(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.service.DeleteMember(c.Request.Context(), id)
	if err != nil {
		switch err.Error() {
		case "member not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "member deleted"})
}

// GetParticipations handles GET /api/meals/participations
func (h *MealHandler) GetParticipations(c *gin.Context) {
	participations, err := h.service.GetParticipations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"participations": participations})
}

// UpdateParticipations handles PUT /api/meals/participations
func (h *MealHandler) UpdateParticipations(c *gin.Context) {
	var req UpdateParticipationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member_id is required"})
		return
	}

	if err := h.service.UpdateParticipations(c.Request.Context(), req.MemberID, req.Breakfast, req.Lunch); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "participations updated"})
}

// GetNextPayer handles GET /api/meals/next-payer
func (h *MealHandler) GetNextPayer(c *gin.Context) {
	result, err := h.service.GetNextPayer(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetPayments handles GET /api/meals/payments?limit=10&offset=0
func (h *MealHandler) GetPayments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	payments, err := h.service.GetPayments(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

// RecordPayment handles POST /api/meals/payments
func (h *MealHandler) RecordPayment(c *gin.Context) {
	var req RecordPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "member_id, meal_type, and date are required"})
		return
	}

	// Validate meal_type
	if req.MealType != "breakfast" && req.MealType != "lunch" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meal type"})
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	payment, err := h.service.RecordPayment(c.Request.Context(), req.MemberID, req.MealType, date)
	if err != nil {
		switch err.Error() {
		case "member not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "member not found"})
		case "invalid meal type":
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid meal type"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payment": payment})
}

// UndoPayment handles DELETE /api/meals/payments/:id
func (h *MealHandler) UndoPayment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	err = h.service.UndoPayment(c.Request.Context(), id)
	if err != nil {
		switch err.Error() {
		case "payment not found":
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		case "can only undo payments within 24 hours":
			c.JSON(http.StatusBadRequest, gin.H{"error": "can only undo payments within 24 hours"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment undone"})
}

// GetStats handles GET /api/meals/stats
func (h *MealHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}
