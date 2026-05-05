package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Newsletter subscriber statuses.
const (
	StatusPending      = "pending"
	StatusActive       = "active"
	StatusUnsubscribed = "unsubscribed"
)

// Sentinel errors for newsletter operations.
var (
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrEmailAlreadyExists = errors.New("email already subscribed")
	ErrSubscriberNotFound = errors.New("subscriber not found")
	ErrTokenNotFound      = errors.New("verification token not found")
	ErrTokenExpired       = errors.New("verification token expired")
)

// emailRegex is a simplified RFC 5322 compliant email validation pattern.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

// NewsletterService defines the interface for newsletter subscription operations.
type NewsletterService interface {
	// Subscribe registers a new email for the newsletter.
	// Returns ErrInvalidEmail if email format is invalid.
	// Returns ErrEmailAlreadyExists if email is already subscribed.
	Subscribe(ctx context.Context, email string) error
	// Unsubscribe marks a subscriber as unsubscribed.
	// Returns ErrSubscriberNotFound if email doesn't exist.
	Unsubscribe(ctx context.Context, email string) error
	// VerifyEmail verifies a subscriber's email using the verification token.
	// Returns ErrTokenNotFound if token doesn't exist.
	VerifyEmail(ctx context.Context, token string) error
}

// newsletterService implements NewsletterService.
type newsletterService struct {
	db *pgxpool.Pool
}

// NewNewsletterService creates a new NewsletterService instance.
func NewNewsletterService(db *pgxpool.Pool) NewsletterService {
	return &newsletterService{
		db: db,
	}
}

// ValidateEmail checks if the given email address is valid according to RFC 5322.
func ValidateEmail(email string) bool {
	if len(email) == 0 || len(email) > 320 {
		return false
	}
	return emailRegex.MatchString(email)
}

// generateVerificationToken creates a cryptographically secure random token (32 bytes hex-encoded).
func generateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// Subscribe registers a new email for the newsletter.
func (s *newsletterService) Subscribe(ctx context.Context, email string) error {
	// Normalize email to lowercase
	email = strings.ToLower(strings.TrimSpace(email))

	// Validate email format
	if !ValidateEmail(email) {
		return ErrInvalidEmail
	}

	// Check if email already exists
	var existingStatus string
	err := s.db.QueryRow(ctx,
		"SELECT status FROM newsletter_subscribers WHERE email = $1", email,
	).Scan(&existingStatus)

	if err == nil {
		// Email exists - handle based on current status
		switch existingStatus {
		case StatusActive:
			return ErrEmailAlreadyExists
		case StatusPending:
			return ErrEmailAlreadyExists
		case StatusUnsubscribed:
			// Re-subscribe: generate new token and update status to pending
			token, tokenErr := generateVerificationToken()
			if tokenErr != nil {
				return tokenErr
			}
			_, updateErr := s.db.Exec(ctx, `
				UPDATE newsletter_subscribers
				SET status = $1, verification_token = $2, subscribed_at = NOW(), unsubscribed_at = NULL
				WHERE email = $3
			`, StatusPending, token, email)
			if updateErr != nil {
				log.Printf("ERROR: Failed to re-subscribe email %s: %v", email, updateErr)
				return fmt.Errorf("failed to re-subscribe: %w", updateErr)
			}
			return nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		log.Printf("ERROR: Failed to check existing subscriber: %v", err)
		return fmt.Errorf("failed to check existing subscriber: %w", err)
	}

	// Generate verification token
	token, err := generateVerificationToken()
	if err != nil {
		return err
	}

	// Insert new subscriber
	_, err = s.db.Exec(ctx, `
		INSERT INTO newsletter_subscribers (email, status, verification_token, subscribed_at, created_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, email, StatusPending, token)
	if err != nil {
		// Handle race condition where email was inserted between our check and insert
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return ErrEmailAlreadyExists
		}
		log.Printf("ERROR: Failed to insert subscriber: %v", err)
		return fmt.Errorf("failed to create subscriber: %w", err)
	}

	return nil
}

// Unsubscribe marks a subscriber as unsubscribed.
func (s *newsletterService) Unsubscribe(ctx context.Context, email string) error {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	if !ValidateEmail(email) {
		return ErrInvalidEmail
	}

	// Update status to unsubscribed
	result, err := s.db.Exec(ctx, `
		UPDATE newsletter_subscribers
		SET status = $1, unsubscribed_at = $2
		WHERE email = $3 AND status != $4
	`, StatusUnsubscribed, time.Now(), email, StatusUnsubscribed)
	if err != nil {
		log.Printf("ERROR: Failed to unsubscribe email %s: %v", email, err)
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	if result.RowsAffected() == 0 {
		// Check if email exists at all
		var exists bool
		checkErr := s.db.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM newsletter_subscribers WHERE email = $1)", email,
		).Scan(&exists)
		if checkErr != nil || !exists {
			return ErrSubscriberNotFound
		}
		// Email exists but already unsubscribed - treat as success
		return nil
	}

	return nil
}

// VerifyEmail verifies a subscriber's email using the verification token.
func (s *newsletterService) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrTokenNotFound
	}

	// Find subscriber by token and update status to active
	result, err := s.db.Exec(ctx, `
		UPDATE newsletter_subscribers
		SET status = $1, verification_token = NULL
		WHERE verification_token = $2 AND status = $3
	`, StatusActive, token, StatusPending)
	if err != nil {
		log.Printf("ERROR: Failed to verify email with token: %v", err)
		return fmt.Errorf("failed to verify email: %w", err)
	}

	if result.RowsAffected() == 0 {
		// Check if token exists but status is not pending
		var status string
		checkErr := s.db.QueryRow(ctx,
			"SELECT status FROM newsletter_subscribers WHERE verification_token = $1", token,
		).Scan(&status)
		if checkErr != nil {
			return ErrTokenNotFound
		}
		// Token exists but subscriber is already active or unsubscribed
		return nil
	}

	return nil
}
