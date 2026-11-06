package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/outbound"
)

type InviteRepository struct {
	pool *Pool
}

func NewInviteRepository(pool *Pool) outbound.InviteRepository {
	return &InviteRepository{pool: pool}
}

const inviteColumns = `id, token_hash, email, display_name, requested_handle, created_by, expires_at,
	claimed_by, claimed_at, revoked_at, created_at`

func (r *InviteRepository) Create(ctx context.Context, inv onboarding.Invite) (*onboarding.Invite, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create invite: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// An expired invite still holds its handle in the partial unique index, so expiry has to be
	// turned into an explicit revocation before the handle can be reused.
	if inv.RequestedHandle != "" {
		if _, err := tx.Exec(ctx, `
			UPDATE artist_invites SET revoked_at = $2
			WHERE LOWER(requested_handle) = LOWER($1)
			  AND claimed_at IS NULL AND revoked_at IS NULL AND expires_at <= $2
		`, inv.RequestedHandle, inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("create invite: release expired handle: %w", err)
		}
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO artist_invites (`+inviteColumns+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING `+inviteColumns,
		inv.ID, inv.TokenHash, inv.Email, inv.DisplayName, inv.RequestedHandle, inv.CreatedBy, inv.ExpiresAt,
		inv.ClaimedBy, inv.ClaimedAt, inv.RevokedAt, inv.CreatedAt)
	saved, err := scanInvite(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("create invite: %w", ErrConflict)
		}
		return nil, fmt.Errorf("create invite: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create invite: commit: %w", err)
	}
	return &saved, nil
}

func (r *InviteRepository) getOne(ctx context.Context, where string, args ...any) (*onboarding.Invite, error) {
	inv, err := scanInvite(r.pool.QueryRow(ctx, `SELECT `+inviteColumns+` FROM artist_invites WHERE `+where, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get invite: %w", err)
	}
	return &inv, nil
}

func (r *InviteRepository) GetByID(ctx context.Context, id uuid.UUID) (*onboarding.Invite, error) {
	return r.getOne(ctx, `id = $1`, id)
}

func (r *InviteRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*onboarding.Invite, error) {
	return r.getOne(ctx, `token_hash = $1`, tokenHash)
}

func (r *InviteRepository) GetLiveByHandle(ctx context.Context, handle string, now time.Time) (*onboarding.Invite, error) {
	return r.getOne(ctx, `LOWER(requested_handle) = LOWER($1) AND requested_handle <> ''
		AND claimed_at IS NULL AND revoked_at IS NULL AND expires_at > $2`, handle, now)
}

func (r *InviteRepository) List(ctx context.Context) ([]onboarding.Invite, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+inviteColumns+` FROM artist_invites ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list invites: %w", err)
	}
	defer rows.Close()

	invites := []onboarding.Invite{}
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, fmt.Errorf("list invites: %w", err)
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func (r *InviteRepository) Claim(ctx context.Context, id, userID uuid.UUID, now time.Time) (*onboarding.Invite, error) {
	// A single conditional UPDATE: of two concurrent claims, only one matches the WHERE clause.
	inv, err := scanInvite(r.pool.QueryRow(ctx, `
		UPDATE artist_invites SET claimed_by = $2, claimed_at = $3
		WHERE id = $1 AND claimed_at IS NULL AND revoked_at IS NULL AND expires_at > $3
		RETURNING `+inviteColumns, id, userID, now))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			if _, getErr := r.GetByID(ctx, id); getErr != nil {
				return nil, getErr // ErrNotFound when the id does not exist at all
			}
			return nil, fmt.Errorf("claim invite: %w", ErrConflict)
		}
		return nil, fmt.Errorf("claim invite: %w", err)
	}
	return &inv, nil
}

func (r *InviteRepository) Revoke(ctx context.Context, id uuid.UUID, now time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE artist_invites SET revoked_at = COALESCE(revoked_at, $2)
		WHERE id = $1 AND claimed_at IS NULL
	`, id, now)
	if err != nil {
		return fmt.Errorf("revoke invite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if _, getErr := r.GetByID(ctx, id); getErr != nil {
			return getErr
		}
		return fmt.Errorf("revoke invite: already claimed: %w", ErrConflict)
	}
	return nil
}

func (r *InviteRepository) Reopen(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE artist_invites SET revoked_at = NULL, claimed_by = NULL, claimed_at = NULL
		WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("reopen invite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanInvite(row scannable) (onboarding.Invite, error) {
	var inv onboarding.Invite
	err := row.Scan(&inv.ID, &inv.TokenHash, &inv.Email, &inv.DisplayName, &inv.RequestedHandle, &inv.CreatedBy,
		&inv.ExpiresAt, &inv.ClaimedBy, &inv.ClaimedAt, &inv.RevokedAt, &inv.CreatedAt)
	return inv, err
}
