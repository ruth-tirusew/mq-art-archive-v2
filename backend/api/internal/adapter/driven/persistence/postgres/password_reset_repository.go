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

type PasswordResetRepository struct {
	pool *Pool
}

func NewPasswordResetRepository(pool *Pool) outbound.PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, token outbound.PasswordResetToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at, used_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.UsedAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("create password reset token: %w", err)
	}
	return nil
}

func (r *PasswordResetRepository) GetByHash(ctx context.Context, tokenHash string) (*outbound.PasswordResetToken, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`, tokenHash)

	var t outbound.PasswordResetToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get password reset token: %w", err)
	}
	return &t, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := r.pool.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL
	`, id, now)
	if err != nil {
		return fmt.Errorf("mark password reset used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
