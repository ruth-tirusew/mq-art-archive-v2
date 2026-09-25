package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/port/outbound"
)

type TelegramLinkRepository struct {
	pool *Pool
}

func NewTelegramLinkRepository(pool *Pool) outbound.TelegramLinkRepository {
	return &TelegramLinkRepository{pool: pool}
}

func (r *TelegramLinkRepository) Create(ctx context.Context, token digest.LinkToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO telegram_link_tokens (token, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`, token.Token, token.UserID, token.CreatedAt, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create telegram link token: %w", err)
	}
	return nil
}

// Consume is a single conditional UPDATE, not a read-then-write, so two concurrent /start
// requests with the same token (unlikely, but the bot has no way to prevent a user from
// tapping the link twice) can't both succeed — only the first commits.
func (r *TelegramLinkRepository) Consume(ctx context.Context, token string) (*digest.LinkToken, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE telegram_link_tokens
		SET consumed_at = NOW()
		WHERE token = $1 AND consumed_at IS NULL AND expires_at > NOW()
		RETURNING token, user_id, created_at, expires_at, consumed_at
	`, token)

	var lt digest.LinkToken
	err := row.Scan(&lt.Token, &lt.UserID, &lt.CreatedAt, &lt.ExpiresAt, &lt.ConsumedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("consume telegram link token: %w", err)
	}
	return &lt, nil
}
