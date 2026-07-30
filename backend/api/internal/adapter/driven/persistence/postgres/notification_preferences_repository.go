package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
)

type NotificationPreferencesRepository struct {
	pool *Pool
}

func NewNotificationPreferencesRepository(pool *Pool) outbound.NotificationPreferencesRepository {
	return &NotificationPreferencesRepository{pool: pool}
}

func (r *NotificationPreferencesRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT user_id, email_on_new_application, email_on_event_sync_summary, newsletter_enabled, updated_at
		FROM user_notification_preferences
		WHERE user_id = $1
	`, userID)

	var prefs identity.NotificationPreferences
	err := row.Scan(
		&prefs.UserID,
		&prefs.EmailOnNewApplication,
		&prefs.EmailOnEventSyncSummary,
		&prefs.NewsletterEnabled,
		&prefs.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get notification preferences: %w", err)
	}
	return &prefs, nil
}

func (r *NotificationPreferencesRepository) Upsert(ctx context.Context, prefs identity.NotificationPreferences) error {
	now := time.Now().UTC()
	if prefs.UpdatedAt.IsZero() {
		prefs.UpdatedAt = now
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_notification_preferences (
			user_id, email_on_new_application, email_on_event_sync_summary, newsletter_enabled, updated_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			email_on_new_application = EXCLUDED.email_on_new_application,
			email_on_event_sync_summary = EXCLUDED.email_on_event_sync_summary,
			newsletter_enabled = EXCLUDED.newsletter_enabled,
			updated_at = EXCLUDED.updated_at
	`, prefs.UserID, prefs.EmailOnNewApplication, prefs.EmailOnEventSyncSummary, prefs.NewsletterEnabled, prefs.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upsert notification preferences: %w", err)
	}
	return nil
}

func (r *NotificationPreferencesRepository) ListEventSummaryRecipients(ctx context.Context) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT u.email
		FROM notification_preferences p
		JOIN users u ON u.id = p.user_id
		WHERE p.email_on_event_sync_summary = TRUE OR p.newsletter_enabled = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	emails := []string{}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}
