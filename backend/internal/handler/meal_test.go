package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/personal-blog/backend/internal/middleware"
	"github.com/personal-blog/backend/internal/service"
)

// --- Mock MealService ---

type mockMealService struct {
	getMembersFunc          func(ctx context.Context) ([]service.MealMember, error)
	createMemberFunc        func(ctx context.Context, name string) (*service.MealMember, error)
	updateMemberFunc        func(ctx context.Context, id int64, name string, isActive bool) (*service.MealMember, error)
	deleteMemberFunc        func(ctx context.Context, id int64) error
	getParticipationsFunc   func(ctx context.Context) ([]service.MealParticipation, error)
	updateParticipationsFunc func(ctx context.Context, memberID int64, breakfast bool, lunch bool) error
	getNextPayerFunc        func(ctx context.Context) (*service.NextPayerResult, error)
	getPaymentsFunc         func(ctx context.Context, limit int, offset int) ([]service.MealPayment, error)
	recordPaymentFunc       func(ctx context.Context, memberID int64, mealType string, date time.Time) (*service.MealPayment, error)
	undoPaymentFunc         func(ctx context.Context, paymentID int64) error
	getStatsFunc            func(ctx context.Context) ([]service.MemberStats, error)
}

func (m *mockMealService) GetMembers(ctx context.Context) ([]service.MealMember, error) {
	if m.getMembersFunc != nil {
		return m.getMembersFunc(ctx)
	}
	return []service.MealMember{}, nil
}

func (m *mockMealService) CreateMember(ctx context.Context, name string) (*service.MealMember, error) {
	if m.createMemberFunc != nil {
		return m.createMemberFunc(ctx, name)
	}
	return &service.MealMember{ID: 1, Name: name, IsActive: true}, nil
}

func (m *mockMealService) UpdateMember(ctx context.Context, id int64, name string, isActive bool) (*service.MealMember, error) {
	if m.updateMemberFunc != nil {
		return m.updateMemberFunc(ctx, id, name, isActive)
	}
	return &service.MealMember{ID: id, Name: name, IsActive: isActive}, nil
}

func (m *mockMealService) DeleteMember(ctx context.Context, id int64) error {
	if m.deleteMemberFunc != nil {
		return m.deleteMemberFunc(ctx, id)
	}
	return nil
}

func (m *mockMealService) GetParticipations(ctx context.Context) ([]service.MealParticipation, error) {
	if m.getParticipationsFunc != nil {
		return m.getParticipationsFunc(ctx)
	}
	return []service.MealParticipation{}, nil
}

func (m *mockMealService) UpdateParticipations(ctx context.Context, memberID int64, breakfast bool, lunch bool) error {
	if m.updateParticipationsFunc != nil {
		return m.updateParticipationsFunc(ctx, memberID, breakfast, lunch)
	}
	return nil
}

func (m *mockMealService) GetNextPayer(ctx context.Context) (*service.NextPayerResult, error) {
	if m.getNextPayerFunc != nil {
		return m.getNextPayerFunc(ctx)
	}
	return &service.NextPayerResult{}, nil
}

func (m *mockMealService) GetPayments(ctx context.Context, limit int, offset int) ([]service.MealPayment, error) {
	if m.getPaymentsFunc != nil {
		return m.getPaymentsFunc(ctx, limit, offset)
	}
	return []service.MealPayment{}, nil
}

func (m *mockMealService) RecordPayment(ctx context.Context, memberID int64, mealType string, date time.Time) (*service.MealPayment, error) {
	if m.recordPaymentFunc != nil {
		return m.recordPaymentFunc(ctx, memberID, mealType, date)
	}
	return &service.MealPayment{ID: 1, MemberID: memberID, MealType: mealType}, nil
}

func (m *mockMealService) UndoPayment(ctx context.Context, paymentID int64) error {
	if m.undoPaymentFunc != nil {
		return m.undoPaymentFunc(ctx, paymentID)
	}
	return nil
}

func (m *mockMealService) GetStats(ctx context.Context) ([]service.MemberStats, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx)
	}
	return []service.MemberStats{}, nil
}

// --- Helper to create router with auth ---

func setupMealRouter(secret string, mock *mockMealService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api")
	mealGroup := api.Group("")
	mealGroup.Use(middleware.MealAuth(secret))
	h := NewMealHandler(mock)
	h.RegisterRoutes(mealGroup)
	return router
}

const testSecret = "test-secret-123"

// =============================================================================
// Auth Middleware Tests
// =============================================================================

func TestMealAuth_ValidSecret(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	req := httptest.NewRequest(http.MethodGet, "/api/meals/members", nil)
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMealAuth_InvalidSecret(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	req := httptest.NewRequest(http.MethodGet, "/api/meals/members", nil)
	req.Header.Set("X-Group-Secret", "wrong-secret")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMealAuth_MissingHeader(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	req := httptest.NewRequest(http.MethodGet, "/api/meals/members", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// =============================================================================
// CreateMember Tests
// =============================================================================

func TestMealHandler_CreateMember_Success(t *testing.T) {
	mock := &mockMealService{
		createMemberFunc: func(ctx context.Context, name string) (*service.MealMember, error) {
			return &service.MealMember{ID: 1, Name: name, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}
	router := setupMealRouter(testSecret, mock)

	body := `{"name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/meals/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestMealHandler_CreateMember_EmptyName(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/meals/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestMealHandler_CreateMember_Duplicate(t *testing.T) {
	mock := &mockMealService{
		createMemberFunc: func(ctx context.Context, name string) (*service.MealMember, error) {
			return nil, fmt.Errorf("member already exists")
		},
	}
	router := setupMealRouter(testSecret, mock)

	body := `{"name":"Alice"}`
	req := httptest.NewRequest(http.MethodPost, "/api/meals/members", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

// =============================================================================
// RecordPayment Tests
// =============================================================================

func TestMealHandler_RecordPayment_Success(t *testing.T) {
	mock := &mockMealService{
		recordPaymentFunc: func(ctx context.Context, memberID int64, mealType string, date time.Time) (*service.MealPayment, error) {
			return &service.MealPayment{ID: 1, MemberID: memberID, MemberName: "Alice", MealType: mealType, PaymentDate: "2026-05-12", CreatedAt: time.Now()}, nil
		},
	}
	router := setupMealRouter(testSecret, mock)

	body := `{"member_id":1,"meal_type":"breakfast","date":"2026-05-12"}`
	req := httptest.NewRequest(http.MethodPost, "/api/meals/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestMealHandler_RecordPayment_InvalidMealType(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	body := `{"member_id":1,"meal_type":"dinner","date":"2026-05-12"}`
	req := httptest.NewRequest(http.MethodPost, "/api/meals/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "invalid meal type" {
		t.Errorf("expected 'invalid meal type' error, got '%v'", resp["error"])
	}
}

func TestMealHandler_RecordPayment_InvalidDate(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	body := `{"member_id":1,"meal_type":"breakfast","date":"not-a-date"}`
	req := httptest.NewRequest(http.MethodPost, "/api/meals/payments", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// =============================================================================
// UndoPayment Tests
// =============================================================================

func TestMealHandler_UndoPayment_Expired(t *testing.T) {
	mock := &mockMealService{
		undoPaymentFunc: func(ctx context.Context, paymentID int64) error {
			return fmt.Errorf("can only undo payments within 24 hours")
		},
	}
	router := setupMealRouter(testSecret, mock)

	req := httptest.NewRequest(http.MethodDelete, "/api/meals/payments/1", nil)
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestMealHandler_UndoPayment_InvalidID(t *testing.T) {
	router := setupMealRouter(testSecret, &mockMealService{})

	req := httptest.NewRequest(http.MethodDelete, "/api/meals/payments/abc", nil)
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// =============================================================================
// DeleteMember Tests
// =============================================================================

func TestMealHandler_DeleteMember_NotFound(t *testing.T) {
	mock := &mockMealService{
		deleteMemberFunc: func(ctx context.Context, id int64) error {
			return fmt.Errorf("member not found")
		},
	}
	router := setupMealRouter(testSecret, mock)

	req := httptest.NewRequest(http.MethodDelete, "/api/meals/members/999", nil)
	req.Header.Set("X-Group-Secret", testSecret)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
