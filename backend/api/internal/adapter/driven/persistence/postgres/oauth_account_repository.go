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

type OAuthAccountRepository struct {
	pool *Pool
}

func NewOAuthAccountRepository(pool *Pool) outbound.OAuthAccountRepository {
	return &OAuthAccountRepository{pool: pool}
}

const oauthAccountColumns = `id, user_id, provider, provider_user_id, email, created_at`

func (r *OAuthAccountRepository) GetByProviderSubject(ctx context.Context, provider, providerUserID string) (*identity.OAuthAccount, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+oauthAccountColumns+`
		FROM oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2
	`, provider, providerUserID)

	account, err := scanOAuthAccount(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get oauth account: %w", err)
	}
	return &account, nil
}

func (r *OAuthAccountRepository) Create(ctx context.Context, account identity.OAuthAccount) error {
	now := time.Now().UTC()
	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}
	if account.CreatedAt.IsZero() {
		account.CreatedAt = now
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, email, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, account.ID, account.UserID, account.Provider, account.ProviderUserID, account.Email, account.CreatedAt)
	if err != nil {
		return fmt.Errorf("create oauth account: %w", err)
	}
	return nil
}

func scanOAuthAccount(row scannable) (identity.OAuthAccount, error) {
	var account identity.OAuthAccount
	err := row.Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.Email,
		&account.CreatedAt,
	)
	return account, err
}
