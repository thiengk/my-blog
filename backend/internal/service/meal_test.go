package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// newTestDBPool creates a PostgreSQL connection pool for testing.
// Tests using this require a running PostgreSQL instance or will be skipped.
func newTestDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use test database URL from environment or default local
	dbURL := "postgres://blog_user:blog_password@localhost:5432/blog_dev?sslmode=disable"

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping test: PostgreSQL not available: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping test: PostgreSQL not reachable: %v", err)
	}

	// Ensure meal tables exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS meal_members (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			deleted_at TIMESTAMP WITH TIME ZONE
		);
		CREATE TABLE IF NOT EXISTS meal_participations (
			id BIGSERIAL PRIMARY KEY,
			member_id BIGINT NOT NULL REFERENCES meal_members(id) ON DELETE CASCADE,
			meal_type VARCHAR(20) NOT NULL CHECK (meal_type IN ('breakfast', 'lunch')),
			is_participating BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(member_id, meal_type)
		);
		CREATE TABLE IF NOT EXISTS meal_payments (
			id BIGSERIAL PRIMARY KEY,
			member_id BIGINT NOT NULL REFERENCES meal_members(id) ON DELETE CASCADE,
			meal_type VARCHAR(20) NOT NULL CHECK (meal_type IN ('breakfast', 'lunch')),
			payment_date DATE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		pool.Close()
		t.Skipf("Skipping test: Failed to create meal tables: %v", err)
	}

	// Clean up test data before each test
	cleanupMealTables(ctx, pool)

	t.Cleanup(func() {
		cleanupMealTables(context.Background(), pool)
		pool.Close()
	})

	return pool
}

func cleanupMealTables(ctx context.Context, pool *pgxpool.Pool) {
	pool.Exec(ctx, "DELETE FROM meal_payments")
	pool.Exec(ctx, "DELETE FROM meal_participations")
	pool.Exec(ctx, "DELETE FROM meal_members")
}

// =============================================================================
// Member CRUD Tests
// =============================================================================

func TestCreateMember_Success(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	member, err := svc.CreateMember(ctx, "Alice")
	if err != nil {
		t.Fatalf("CreateMember failed: %v", err)
	}

	if member.Name != "Alice" {
		t.Errorf("expected name = Alice, got %s", member.Name)
	}
	if !member.IsActive {
		t.Error("expected new member to be active")
	}
	if member.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestCreateMember_DuplicateName(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	_, err := svc.CreateMember(ctx, "Alice")
	if err != nil {
		t.Fatalf("first CreateMember failed: %v", err)
	}

	_, err = svc.CreateMember(ctx, "Alice")
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
	if err.Error() != "member already exists" {
		t.Errorf("expected 'member already exists' error, got: %v", err)
	}
}

func TestCreateMember_EmptyName(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	_, err := svc.CreateMember(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if err.Error() != "name is required" {
		t.Errorf("expected 'name is required' error, got: %v", err)
	}
}

func TestCreateMember_AutoCreatesParticipations(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	member, err := svc.CreateMember(ctx, "Bob")
	if err != nil {
		t.Fatalf("CreateMember failed: %v", err)
	}

	participations, err := svc.GetParticipations(ctx)
	if err != nil {
		t.Fatalf("GetParticipations failed: %v", err)
	}

	// Should have 2 participations (breakfast + lunch)
	count := 0
	for _, p := range participations {
		if p.MemberID == member.ID {
			count++
			if !p.IsParticipating {
				t.Errorf("expected default participation to be true for %s", p.MealType)
			}
		}
	}
	if count != 2 {
		t.Errorf("expected 2 participations for new member, got %d", count)
	}
}

func TestGetMembers_ExcludesDeleted(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	_, err := svc.CreateMember(ctx, "Alice")
	if err != nil {
		t.Fatalf("CreateMember Alice failed: %v", err)
	}

	bob, err := svc.CreateMember(ctx, "Bob")
	if err != nil {
		t.Fatalf("CreateMember Bob failed: %v", err)
	}

	// Delete Bob
	err = svc.DeleteMember(ctx, bob.ID)
	if err != nil {
		t.Fatalf("DeleteMember failed: %v", err)
	}

	members, err := svc.GetMembers(ctx)
	if err != nil {
		t.Fatalf("GetMembers failed: %v", err)
	}

	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].Name != "Alice" {
		t.Errorf("expected Alice, got %s", members[0].Name)
	}
}

func TestUpdateMember_Success(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	member, err := svc.CreateMember(ctx, "Alice")
	if err != nil {
		t.Fatalf("CreateMember failed: %v", err)
	}

	updated, err := svc.UpdateMember(ctx, member.ID, "Alice Updated", false)
	if err != nil {
		t.Fatalf("UpdateMember failed: %v", err)
	}

	if updated.Name != "Alice Updated" {
		t.Errorf("expected name = 'Alice Updated', got %s", updated.Name)
	}
	if updated.IsActive {
		t.Error("expected member to be inactive")
	}
}

func TestDeleteMember_NotFound(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	err := svc.DeleteMember(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent member")
	}
	if err.Error() != "member not found" {
		t.Errorf("expected 'member not found' error, got: %v", err)
	}
}

// =============================================================================
// Scheduling Algorithm Tests
// =============================================================================

func TestGetNextPayer_SingleMember(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	_, err := svc.CreateMember(ctx, "Alice")
	if err != nil {
		t.Fatalf("CreateMember failed: %v", err)
	}

	result, err := svc.GetNextPayer(ctx)
	if err != nil {
		t.Fatalf("GetNextPayer failed: %v", err)
	}

	if result.Breakfast == nil {
		t.Fatal("expected breakfast payer, got nil")
	}
	if result.Breakfast.MemberName != "Alice" {
		t.Errorf("expected Alice, got %s", result.Breakfast.MemberName)
	}
	if result.Breakfast.PaymentCount != 0 {
		t.Errorf("expected payment count = 0, got %d", result.Breakfast.PaymentCount)
	}
}

func TestGetNextPayer_PicksLowestCount(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")
	_, _ = svc.CreateMember(ctx, "Bob")

	// Alice pays twice for breakfast
	svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now())
	svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now())

	result, err := svc.GetNextPayer(ctx)
	if err != nil {
		t.Fatalf("GetNextPayer failed: %v", err)
	}

	// Bob should be next (0 payments vs Alice's 2)
	if result.Breakfast == nil {
		t.Fatal("expected breakfast payer, got nil")
	}
	if result.Breakfast.MemberName != "Bob" {
		t.Errorf("expected Bob (0 payments), got %s (%d payments)",
			result.Breakfast.MemberName, result.Breakfast.PaymentCount)
	}
}

func TestGetNextPayer_TieBreakByLastPaid(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")
	bob, _ := svc.CreateMember(ctx, "Bob")

	// Both pay once, but Alice paid more recently
	svc.RecordPayment(ctx, bob.ID, "breakfast", time.Now().AddDate(0, 0, -5))
	svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now().AddDate(0, 0, -1))

	result, err := svc.GetNextPayer(ctx)
	if err != nil {
		t.Fatalf("GetNextPayer failed: %v", err)
	}

	// Bob should be next (paid longer ago)
	if result.Breakfast == nil {
		t.Fatal("expected breakfast payer, got nil")
	}
	if result.Breakfast.MemberName != "Bob" {
		t.Errorf("expected Bob (paid 5 days ago), got %s", result.Breakfast.MemberName)
	}
}

func TestGetNextPayer_NewMemberPrioritized(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")
	svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now())

	// Add new member - should be prioritized (0 payments, NULL last_paid)
	_, _ = svc.CreateMember(ctx, "Charlie")

	result, err := svc.GetNextPayer(ctx)
	if err != nil {
		t.Fatalf("GetNextPayer failed: %v", err)
	}

	if result.Breakfast == nil {
		t.Fatal("expected breakfast payer, got nil")
	}
	if result.Breakfast.MemberName != "Charlie" {
		t.Errorf("expected Charlie (new member, 0 payments), got %s", result.Breakfast.MemberName)
	}
}

func TestGetNextPayer_NoParticipants(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")

	// Remove Alice from breakfast
	svc.UpdateParticipations(ctx, alice.ID, false, true)

	result, err := svc.GetNextPayer(ctx)
	if err != nil {
		t.Fatalf("GetNextPayer failed: %v", err)
	}

	// No breakfast participants
	if result.Breakfast != nil {
		t.Errorf("expected nil breakfast payer, got %s", result.Breakfast.MemberName)
	}
	// Alice still participates in lunch
	if result.Lunch == nil {
		t.Fatal("expected lunch payer, got nil")
	}
	if result.Lunch.MemberName != "Alice" {
		t.Errorf("expected Alice for lunch, got %s", result.Lunch.MemberName)
	}
}

// =============================================================================
// Payment Tests
// =============================================================================

func TestRecordPayment_Success(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")

	payment, err := svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now())
	if err != nil {
		t.Fatalf("RecordPayment failed: %v", err)
	}

	if payment.MemberName != "Alice" {
		t.Errorf("expected member name = Alice, got %s", payment.MemberName)
	}
	if payment.MealType != "breakfast" {
		t.Errorf("expected meal type = breakfast, got %s", payment.MealType)
	}
}

func TestRecordPayment_InvalidMealType(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")

	_, err := svc.RecordPayment(ctx, alice.ID, "dinner", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid meal type")
	}
	if err.Error() != "invalid meal type" {
		t.Errorf("expected 'invalid meal type' error, got: %v", err)
	}
}

func TestRecordPayment_MemberNotFound(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	_, err := svc.RecordPayment(ctx, 99999, "breakfast", time.Now())
	if err == nil {
		t.Fatal("expected error for non-existent member")
	}
	if err.Error() != "member not found" {
		t.Errorf("expected 'member not found' error, got: %v", err)
	}
}

func TestUndoPayment_WithinWindow(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")
	payment, _ := svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now())

	err := svc.UndoPayment(ctx, payment.ID)
	if err != nil {
		t.Fatalf("UndoPayment failed: %v", err)
	}

	// Verify payment is gone
	payments, _ := svc.GetPayments(ctx, 10, 0)
	if len(payments) != 0 {
		t.Errorf("expected 0 payments after undo, got %d", len(payments))
	}
}

func TestUndoPayment_NotFound(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	err := svc.UndoPayment(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent payment")
	}
	if err.Error() != "payment not found" {
		t.Errorf("expected 'payment not found' error, got: %v", err)
	}
}

// =============================================================================
// Stats Tests
// =============================================================================

func TestGetStats_MultipleMembers(t *testing.T) {
	pool := newTestDBPool(t)
	ctx := context.Background()
	svc := NewMealService(pool)

	alice, _ := svc.CreateMember(ctx, "Alice")
	bob, _ := svc.CreateMember(ctx, "Bob")

	// Alice: 2 breakfast, 1 lunch
	svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now())
	svc.RecordPayment(ctx, alice.ID, "breakfast", time.Now().AddDate(0, 0, -1))
	svc.RecordPayment(ctx, alice.ID, "lunch", time.Now())

	// Bob: 1 breakfast, 2 lunch
	svc.RecordPayment(ctx, bob.ID, "breakfast", time.Now())
	svc.RecordPayment(ctx, bob.ID, "lunch", time.Now())
	svc.RecordPayment(ctx, bob.ID, "lunch", time.Now().AddDate(0, 0, -1))

	stats, err := svc.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf("expected 2 stats entries, got %d", len(stats))
	}

	// Stats are ordered by name ASC
	// Alice first
	if stats[0].MemberName != "Alice" {
		t.Errorf("expected first stat = Alice, got %s", stats[0].MemberName)
	}
	if stats[0].BreakfastCount != 2 {
		t.Errorf("Alice breakfast: expected 2, got %d", stats[0].BreakfastCount)
	}
	if stats[0].LunchCount != 1 {
		t.Errorf("Alice lunch: expected 1, got %d", stats[0].LunchCount)
	}
	if stats[0].TotalCount != 3 {
		t.Errorf("Alice total: expected 3, got %d", stats[0].TotalCount)
	}

	// Bob second
	if stats[1].MemberName != "Bob" {
		t.Errorf("expected second stat = Bob, got %s", stats[1].MemberName)
	}
	if stats[1].BreakfastCount != 1 {
		t.Errorf("Bob breakfast: expected 1, got %d", stats[1].BreakfastCount)
	}
	if stats[1].LunchCount != 2 {
		t.Errorf("Bob lunch: expected 2, got %d", stats[1].LunchCount)
	}
	if stats[1].TotalCount != 3 {
		t.Errorf("Bob total: expected 3, got %d", stats[1].TotalCount)
	}
}
