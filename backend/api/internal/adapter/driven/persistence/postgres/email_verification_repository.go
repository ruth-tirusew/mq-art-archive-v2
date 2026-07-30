package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/port/outbound"
)

type EmailVerificationRepository struct{ pool *Pool }

func NewEmailVerificationRepository(pool *Pool) outbound.EmailVerificationRepository {
	return &EmailVerificationRepository{pool: pool}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, token outbound.EmailVerificationToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO email_verification_tokens (id,user_id,token_hash,expires_at,created_at)
		VALUES ($1,$2,$3,$4,$5)`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("create email verification token: %w", err)
	}
	return nil
}

func (r *EmailVerificationRepository) Consume(ctx context.Context, tokenHash string, now time.Time) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)
	var userID uuid.UUID
	err = tx.QueryRow(ctx, `DELETE FROM email_verification_tokens
		WHERE token_hash=$1 AND expires_at>$2 RETURNING user_id`, tokenHash, now).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("consume email verification token: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE users SET email_verified_at=$2, updated_at=$2 WHERE id=$1`, userID, now); err != nil {
		return uuid.Nil, err
	}
	return userID, tx.Commit(ctx)
}
