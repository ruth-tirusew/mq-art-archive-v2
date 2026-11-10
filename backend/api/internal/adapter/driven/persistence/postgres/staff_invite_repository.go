package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
)

type StaffInviteRepository struct {
	pool *Pool
}

func NewStaffInviteRepository(pool *Pool) outbound.StaffInviteRepository {
	return &StaffInviteRepository{pool: pool}
}

const staffInviteColumns = `id, token_hash, email, role, created_by, expires_at, claimed_by, claimed_at, revoked_at, created_at`

func (r *StaffInviteRepository) Create(ctx context.Context, inv identity.StaffInvite) (*identity.StaffInvite, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create staff invite: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// One live invite per address: whatever is still unclaimed (a resend, or an expired one that still
	// holds the unique index) is revoked so the new link is the only one that works.
	if _, err := tx.Exec(ctx, `
		UPDATE staff_invites SET revoked_at = $2
		WHERE LOWER(email) = LOWER($1) AND claimed_at IS NULL AND revoked_at IS NULL
	`, inv.Email, inv.CreatedAt); err != nil {
		return nil, fmt.Errorf("create staff invite: replace the previous one: %w", err)
	}

	saved, err := scanStaffInvite(tx.QueryRow(ctx, `
		INSERT INTO staff_invites (`+staffInviteColumns+`) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+staffInviteColumns,
		inv.ID, inv.TokenHash, inv.Email, string(inv.Role), inv.CreatedBy, inv.ExpiresAt, inv.ClaimedBy, inv.ClaimedAt, inv.RevokedAt, inv.CreatedAt))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("create staff invite: %w", ErrConflict)
		}
		return nil, fmt.Errorf("create staff invite: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create staff invite: commit: %w", err)
	}
	return &saved, nil
}

func (r *StaffInviteRepository) getOne(ctx context.Context, where string, args ...any) (*identity.StaffInvite, error) {
	inv, err := scanStaffInvite(r.pool.QueryRow(ctx, `SELECT `+staffInviteColumns+` FROM staff_invites WHERE `+where, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get staff invite: %w", err)
	}
	return &inv, nil
}

func (r *StaffInviteRepository) GetByID(ctx context.Context, id uuid.UUID) (*identity.StaffInvite, error) {
	return r.getOne(ctx, `id = $1`, id)
}

func (r *StaffInviteRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*identity.StaffInvite, error) {
	return r.getOne(ctx, `token_hash = $1`, tokenHash)
}

func (r *StaffInviteRepository) List(ctx context.Context, limit, offset int) ([]identity.StaffInvite, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM staff_invites`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count staff invites: %w", err)
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+staffInviteColumns+` FROM staff_invites ORDER BY created_at DESC, id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list staff invites: %w", err)
	}
	defer rows.Close()
	out := []identity.StaffInvite{}
	for rows.Next() {
		inv, err := scanStaffInvite(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("list staff invites: %w", err)
		}
		out = append(out, inv)
	}
	return out, total, rows.Err()
}

func (r *StaffInviteRepository) Accept(ctx context.Context, inviteID uuid.UUID, user identity.User, passwordHash string, now time.Time) (*identity.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("accept staff invite: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// The account first, so the invite's claimed_by can point at it; the claim then decides whether it stays.
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, email, role, created_at, updated_at, password_hash, email_verified_at)
		VALUES ($1, $2, $3, $4, $4, $5, $6)
	`, user.ID, user.Email, string(user.Role), now, passwordHash, user.EmailVerifiedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("accept staff invite: address already has an account: %w", ErrConflict)
		}
		return nil, fmt.Errorf("accept staff invite: create account: %w", err)
	}

	// A single conditional UPDATE: a revoked, expired or already used invite matches no row, and the
	// whole transaction (account included) is rolled back.
	tag, err := tx.Exec(ctx, `
		UPDATE staff_invites SET claimed_by = $2, claimed_at = $3
		WHERE id = $1 AND claimed_at IS NULL AND revoked_at IS NULL AND expires_at > $3
	`, inviteID, user.ID, now)
	if err != nil {
		return nil, fmt.Errorf("accept staff invite: claim: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("accept staff invite: invite is no longer live: %w", ErrConflict)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("accept staff invite: commit: %w", err)
	}
	user.CreatedAt, user.UpdatedAt = now, now
	return &user, nil
}

func (r *StaffInviteRepository) Revoke(ctx context.Context, id uuid.UUID, now time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE staff_invites SET revoked_at = COALESCE(revoked_at, $2) WHERE id = $1 AND claimed_at IS NULL
	`, id, now)
	if err != nil {
		return fmt.Errorf("revoke staff invite: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if _, getErr := r.GetByID(ctx, id); getErr != nil {
			return getErr
		}
		return fmt.Errorf("revoke staff invite: already claimed: %w", ErrConflict)
	}
	return nil
}

func scanStaffInvite(row scannable) (identity.StaffInvite, error) {
	var inv identity.StaffInvite
	var role string
	err := row.Scan(&inv.ID, &inv.TokenHash, &inv.Email, &role, &inv.CreatedBy, &inv.ExpiresAt, &inv.ClaimedBy, &inv.ClaimedAt, &inv.RevokedAt, &inv.CreatedAt)
	inv.Role = identity.Role(role)
	return inv, err
}
