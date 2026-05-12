package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MealMember represents a team member in the meal scheduler.
type MealMember struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// MealParticipation represents a member's participation in a meal type.
type MealParticipation struct {
	ID              int64  `json:"id"`
	MemberID        int64  `json:"member_id"`
	MemberName      string `json:"member_name"`
	MealType        string `json:"meal_type"`
	IsParticipating bool   `json:"is_participating"`
}

// MealPayment represents a recorded payment for a meal.
type MealPayment struct {
	ID          int64     `json:"id"`
	MemberID    int64     `json:"member_id"`
	MemberName  string    `json:"member_name"`
	MealType    string    `json:"meal_type"`
	PaymentDate string    `json:"payment_date"`
	CreatedAt   time.Time `json:"created_at"`
}

// NextPayer represents the next person to pay for a meal type.
type NextPayer struct {
	MemberID     int64      `json:"member_id"`
	MemberName   string     `json:"member_name"`
	PaymentCount int64      `json:"payment_count"`
	LastPaidAt   *time.Time `json:"last_paid_at,omitempty"`
}

// NextPayerResult holds the next payer for each meal type.
type NextPayerResult struct {
	Breakfast *NextPayer `json:"breakfast"`
	Lunch     *NextPayer `json:"lunch"`
}

// MemberStats holds payment statistics for a member.
type MemberStats struct {
	MemberID       int64  `json:"member_id"`
	MemberName     string `json:"member_name"`
	BreakfastCount int64  `json:"breakfast_count"`
	LunchCount     int64  `json:"lunch_count"`
	TotalCount     int64  `json:"total_count"`
}

// MealService defines the business logic for meal payment scheduling.
type MealService interface {
	// Members
	GetMembers(ctx context.Context) ([]MealMember, error)
	CreateMember(ctx context.Context, name string) (*MealMember, error)
	UpdateMember(ctx context.Context, id int64, name string, isActive bool) (*MealMember, error)
	DeleteMember(ctx context.Context, id int64) error

	// Participations
	GetParticipations(ctx context.Context) ([]MealParticipation, error)
	UpdateParticipations(ctx context.Context, memberID int64, breakfast bool, lunch bool) error

	// Scheduling
	GetNextPayer(ctx context.Context) (*NextPayerResult, error)

	// Payments
	GetPayments(ctx context.Context, limit int, offset int) ([]MealPayment, error)
	RecordPayment(ctx context.Context, memberID int64, mealType string, date time.Time) (*MealPayment, error)
	UndoPayment(ctx context.Context, paymentID int64) error

	// Stats
	GetStats(ctx context.Context) ([]MemberStats, error)
}

// mealService implements MealService.
type mealService struct {
	db *pgxpool.Pool
}

// NewMealService creates a new MealService instance.
func NewMealService(db *pgxpool.Pool) MealService {
	return &mealService{db: db}
}

// GetMembers returns all active (non-deleted) members.
func (s *mealService) GetMembers(ctx context.Context) ([]MealMember, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, name, is_active, created_at, updated_at
		FROM meal_members
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []MealMember
	for rows.Next() {
		var m MealMember
		if err := rows.Scan(&m.ID, &m.Name, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, m)
	}
	if members == nil {
		members = []MealMember{}
	}
	return members, nil
}

// CreateMember creates a new member and auto-creates participation records.
func (s *mealService) CreateMember(ctx context.Context, name string) (*MealMember, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var member MealMember
	err = tx.QueryRow(ctx, `
		INSERT INTO meal_members (name, is_active, created_at, updated_at)
		VALUES ($1, true, NOW(), NOW())
		RETURNING id, name, is_active, created_at, updated_at
	`, name).Scan(&member.ID, &member.Name, &member.IsActive, &member.CreatedAt, &member.UpdatedAt)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("member already exists")
		}
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	// Auto-create participation records (default: participating in both meals)
	_, err = tx.Exec(ctx, `
		INSERT INTO meal_participations (member_id, meal_type, is_participating, created_at, updated_at)
		VALUES ($1, 'breakfast', true, NOW(), NOW()),
		       ($1, 'lunch', true, NOW(), NOW())
	`, member.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create participations: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &member, nil
}

// UpdateMember updates a member's name and active status.
func (s *mealService) UpdateMember(ctx context.Context, id int64, name string, isActive bool) (*MealMember, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	var member MealMember
	err := s.db.QueryRow(ctx, `
		UPDATE meal_members
		SET name = $2, is_active = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, name, is_active, created_at, updated_at
	`, id, name, isActive).Scan(&member.ID, &member.Name, &member.IsActive, &member.CreatedAt, &member.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("member not found")
		}
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("member already exists")
		}
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	return &member, nil
}

// DeleteMember performs a soft delete on a member.
func (s *mealService) DeleteMember(ctx context.Context, id int64) error {
	result, err := s.db.Exec(ctx, `
		UPDATE meal_members
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found")
	}
	return nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	return err != nil && contains(err.Error(), "23505")
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetParticipations returns all participation records with member names.
func (s *mealService) GetParticipations(ctx context.Context) ([]MealParticipation, error) {
	rows, err := s.db.Query(ctx, `
		SELECT mp.id, mp.member_id, m.name, mp.meal_type, mp.is_participating
		FROM meal_participations mp
		JOIN meal_members m ON m.id = mp.member_id
		WHERE m.deleted_at IS NULL
		ORDER BY m.name ASC, mp.meal_type ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query participations: %w", err)
	}
	defer rows.Close()

	var participations []MealParticipation
	for rows.Next() {
		var p MealParticipation
		if err := rows.Scan(&p.ID, &p.MemberID, &p.MemberName, &p.MealType, &p.IsParticipating); err != nil {
			return nil, fmt.Errorf("failed to scan participation: %w", err)
		}
		participations = append(participations, p)
	}
	if participations == nil {
		participations = []MealParticipation{}
	}
	return participations, nil
}

// UpdateParticipations updates a member's participation for breakfast and lunch.
func (s *mealService) UpdateParticipations(ctx context.Context, memberID int64, breakfast bool, lunch bool) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert breakfast participation
	_, err = tx.Exec(ctx, `
		INSERT INTO meal_participations (member_id, meal_type, is_participating, created_at, updated_at)
		VALUES ($1, 'breakfast', $2, NOW(), NOW())
		ON CONFLICT (member_id, meal_type)
		DO UPDATE SET is_participating = $2, updated_at = NOW()
	`, memberID, breakfast)
	if err != nil {
		return fmt.Errorf("failed to update breakfast participation: %w", err)
	}

	// Upsert lunch participation
	_, err = tx.Exec(ctx, `
		INSERT INTO meal_participations (member_id, meal_type, is_participating, created_at, updated_at)
		VALUES ($1, 'lunch', $2, NOW(), NOW())
		ON CONFLICT (member_id, meal_type)
		DO UPDATE SET is_participating = $2, updated_at = NOW()
	`, memberID, lunch)
	if err != nil {
		return fmt.Errorf("failed to update lunch participation: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetNextPayer returns the next person to pay for each meal type.
// Algorithm: pick the member with fewest payments; tie-break by longest since last payment; then by ID.
func (s *mealService) GetNextPayer(ctx context.Context) (*NextPayerResult, error) {
	result := &NextPayerResult{}

	breakfast, err := s.getNextPayerForMeal(ctx, "breakfast")
	if err != nil {
		return nil, fmt.Errorf("failed to get next breakfast payer: %w", err)
	}
	result.Breakfast = breakfast

	lunch, err := s.getNextPayerForMeal(ctx, "lunch")
	if err != nil {
		return nil, fmt.Errorf("failed to get next lunch payer: %w", err)
	}
	result.Lunch = lunch

	return result, nil
}

// getNextPayerForMeal returns the next payer for a specific meal type.
func (s *mealService) getNextPayerForMeal(ctx context.Context, mealType string) (*NextPayer, error) {
	var payer NextPayer
	err := s.db.QueryRow(ctx, `
		SELECT m.id, m.name,
		       COALESCE(COUNT(p.id), 0) AS payment_count,
		       MAX(p.payment_date) AS last_paid_at
		FROM meal_members m
		JOIN meal_participations mp ON mp.member_id = m.id
		LEFT JOIN meal_payments p ON p.member_id = m.id AND p.meal_type = $1
		WHERE m.is_active = true
		  AND m.deleted_at IS NULL
		  AND mp.meal_type = $1
		  AND mp.is_participating = true
		GROUP BY m.id, m.name
		ORDER BY payment_count ASC, last_paid_at ASC NULLS FIRST, m.id ASC
		LIMIT 1
	`, mealType).Scan(&payer.MemberID, &payer.MemberName, &payer.PaymentCount, &payer.LastPaidAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No participants for this meal
		}
		return nil, fmt.Errorf("failed to query next payer: %w", err)
	}

	return &payer, nil
}

// GetPayments returns payment history with pagination, newest first.
func (s *mealService) GetPayments(ctx context.Context, limit int, offset int) ([]MealPayment, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(ctx, `
		SELECT p.id, p.member_id, m.name, p.meal_type, p.payment_date, p.created_at
		FROM meal_payments p
		JOIN meal_members m ON m.id = p.member_id
		ORDER BY p.payment_date DESC, p.created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query payments: %w", err)
	}
	defer rows.Close()

	var payments []MealPayment
	for rows.Next() {
		var p MealPayment
		var paymentDate time.Time
		if err := rows.Scan(&p.ID, &p.MemberID, &p.MemberName, &p.MealType, &paymentDate, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		p.PaymentDate = paymentDate.Format("2006-01-02")
		payments = append(payments, p)
	}
	if payments == nil {
		payments = []MealPayment{}
	}
	return payments, nil
}

// RecordPayment records a payment for a member.
func (s *mealService) RecordPayment(ctx context.Context, memberID int64, mealType string, date time.Time) (*MealPayment, error) {
	if mealType != "breakfast" && mealType != "lunch" {
		return nil, fmt.Errorf("invalid meal type")
	}

	// Verify member exists and is active
	var memberName string
	err := s.db.QueryRow(ctx, `
		SELECT name FROM meal_members WHERE id = $1 AND deleted_at IS NULL
	`, memberID).Scan(&memberName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("member not found")
		}
		return nil, fmt.Errorf("failed to verify member: %w", err)
	}

	var payment MealPayment
	var paymentDate time.Time
	err = s.db.QueryRow(ctx, `
		INSERT INTO meal_payments (member_id, meal_type, payment_date, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, member_id, meal_type, payment_date, created_at
	`, memberID, mealType, date).Scan(&payment.ID, &payment.MemberID, &payment.MealType, &paymentDate, &payment.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}
	payment.MemberName = memberName
	payment.PaymentDate = paymentDate.Format("2006-01-02")

	return &payment, nil
}

// UndoPayment deletes a payment if it was created within the last 24 hours.
func (s *mealService) UndoPayment(ctx context.Context, paymentID int64) error {
	var createdAt time.Time
	err := s.db.QueryRow(ctx, `
		SELECT created_at FROM meal_payments WHERE id = $1
	`, paymentID).Scan(&createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("payment not found")
		}
		return fmt.Errorf("failed to query payment: %w", err)
	}

	// Check if within 24 hours
	if time.Since(createdAt) > 24*time.Hour {
		return fmt.Errorf("can only undo payments within 24 hours")
	}

	_, err = s.db.Exec(ctx, `DELETE FROM meal_payments WHERE id = $1`, paymentID)
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	return nil
}

// GetStats returns payment statistics for all active members.
func (s *mealService) GetStats(ctx context.Context) ([]MemberStats, error) {
	rows, err := s.db.Query(ctx, `
		SELECT m.id, m.name,
		       COALESCE(SUM(CASE WHEN p.meal_type = 'breakfast' THEN 1 ELSE 0 END), 0) AS breakfast_count,
		       COALESCE(SUM(CASE WHEN p.meal_type = 'lunch' THEN 1 ELSE 0 END), 0) AS lunch_count,
		       COALESCE(COUNT(p.id), 0) AS total_count
		FROM meal_members m
		LEFT JOIN meal_payments p ON p.member_id = m.id
		WHERE m.deleted_at IS NULL AND m.is_active = true
		GROUP BY m.id, m.name
		ORDER BY m.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	var stats []MemberStats
	for rows.Next() {
		var s MemberStats
		if err := rows.Scan(&s.MemberID, &s.MemberName, &s.BreakfastCount, &s.LunchCount, &s.TotalCount); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		stats = append(stats, s)
	}
	if stats == nil {
		stats = []MemberStats{}
	}
	return stats, nil
}
