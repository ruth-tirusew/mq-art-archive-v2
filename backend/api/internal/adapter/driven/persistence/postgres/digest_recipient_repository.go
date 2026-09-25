package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/port/outbound"
)

type DigestRecipientRepository struct {
	pool *Pool
}

func NewDigestRecipientRepository(pool *Pool) outbound.DigestRecipientRepository {
	return &DigestRecipientRepository{pool: pool}
}

func (r *DigestRecipientRepository) ListDigestRecipients(ctx context.Context) ([]digest.Recipient, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.email, p.newsletter_enabled, p.telegram_chat_id
		FROM user_notification_preferences p
		JOIN users u ON u.id = p.user_id
		WHERE p.newsletter_enabled = TRUE OR p.telegram_chat_id IS NOT NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("list digest recipients: %w", err)
	}
	defer rows.Close()

	recipients := []digest.Recipient{}
	for rows.Next() {
		var rec digest.Recipient
		var userID uuid.UUID
		if err := rows.Scan(&userID, &rec.Email, &rec.EmailOptedIn, &rec.TelegramChatID); err != nil {
			return nil, err
		}
		rec.UserID = userID
		recipients = append(recipients, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list digest recipients: %w", err)
	}
	return recipients, nil
}
