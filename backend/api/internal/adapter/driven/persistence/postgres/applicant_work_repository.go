package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/outbound"
)

type ApplicantWorkRepository struct {
	pool *Pool
}

func NewApplicantWorkRepository(pool *Pool) outbound.ApplicantWorkRepository {
	return &ApplicantWorkRepository{pool: pool}
}

const applicantWorkColumns = `w.id, w.user_id, w.media_asset_id, m.secure_url, m.blurhash, w.title, w.medium, w.year, w.position, w.created_at`

func (r *ApplicantWorkRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]onboarding.ApplicantWork, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+applicantWorkColumns+`
		FROM applicant_works w JOIN media_assets m ON m.id = w.media_asset_id
		WHERE w.user_id = $1 AND w.kind = 'work' ORDER BY w.position, w.created_at, w.id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list applicant works: %w", err)
	}
	defer rows.Close()

	works := []onboarding.ApplicantWork{}
	for rows.Next() {
		w, err := scanApplicantWork(rows)
		if err != nil {
			return nil, fmt.Errorf("list applicant works: %w", err)
		}
		works = append(works, w)
	}
	return works, rows.Err()
}

func (r *ApplicantWorkRepository) Add(ctx context.Context, work onboarding.ApplicantWork) (*onboarding.ApplicantWork, error) {
	// One statement: the image must exist AND belong to this user, and the position goes after
	// their last work. A work for someone else's image matches no row and reads as "not found".
	row := r.pool.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO applicant_works (id, user_id, media_asset_id, title, medium, year, position, created_at)
			SELECT $1, $2, m.id, $4, $5, $6,
			       COALESCE((SELECT MAX(position) + 1 FROM applicant_works WHERE user_id = $2 AND kind = 'work'), 0), $7
			FROM media_assets m WHERE m.id = $3 AND m.owner_user_id = $2
			RETURNING *
		)
		SELECT w.id, w.user_id, w.media_asset_id, m.secure_url, m.blurhash, w.title, w.medium, w.year, w.position, w.created_at
		FROM inserted w JOIN media_assets m ON m.id = w.media_asset_id
	`, work.ID, work.UserID, work.MediaAssetID, work.Title, work.Medium, work.Year, work.CreatedAt)

	saved, err := scanApplicantWork(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("add applicant work: %w", ErrConflict)
		}
		return nil, fmt.Errorf("add applicant work: %w", err)
	}
	return &saved, nil
}

func (r *ApplicantWorkRepository) Delete(ctx context.Context, userID, workID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM applicant_works WHERE id = $1 AND user_id = $2`, workID, userID)
	if err != nil {
		return fmt.Errorf("delete applicant work: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanApplicantWork(row scannable) (onboarding.ApplicantWork, error) {
	var w onboarding.ApplicantWork
	err := row.Scan(&w.ID, &w.UserID, &w.MediaAssetID, &w.URL, &w.Blurhash, &w.Title, &w.Medium, &w.Year, &w.Position, &w.CreatedAt)
	return w, err
}

func (r *ApplicantWorkRepository) GetPortrait(ctx context.Context, userID uuid.UUID) (*onboarding.ApplicantPortrait, error) {
	var p onboarding.ApplicantPortrait
	err := r.pool.QueryRow(ctx, `
		SELECT w.media_asset_id, m.secure_url, m.blurhash
		FROM applicant_works w JOIN media_assets m ON m.id = w.media_asset_id
		WHERE w.user_id = $1 AND w.kind = 'portrait'
	`, userID).Scan(&p.MediaAssetID, &p.URL, &p.Blurhash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get applicant portrait: %w", err)
	}
	return &p, nil
}

func (r *ApplicantWorkRepository) SetPortrait(ctx context.Context, userID, mediaAssetID uuid.UUID) (*onboarding.ApplicantPortrait, error) {
	// One statement, and only for an image that exists and is theirs. The unique index allows a
	// single portrait per user, so a second one replaces the first.
	var p onboarding.ApplicantPortrait
	err := r.pool.QueryRow(ctx, `
		WITH saved AS (
			INSERT INTO applicant_works (id, user_id, media_asset_id, title, kind, position, created_at)
			SELECT gen_random_uuid(), $1, m.id, '', 'portrait', 0, NOW()
			FROM media_assets m WHERE m.id = $2 AND m.owner_user_id = $1
			ON CONFLICT (user_id) WHERE kind = 'portrait'
			DO UPDATE SET media_asset_id = EXCLUDED.media_asset_id, created_at = EXCLUDED.created_at
			RETURNING media_asset_id
		)
		SELECT s.media_asset_id, m.secure_url, m.blurhash FROM saved s JOIN media_assets m ON m.id = s.media_asset_id
	`, userID, mediaAssetID).Scan(&p.MediaAssetID, &p.URL, &p.Blurhash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("set applicant portrait: %w", ErrConflict)
		}
		return nil, fmt.Errorf("set applicant portrait: %w", err)
	}
	return &p, nil
}

func (r *ApplicantWorkRepository) DeletePortrait(ctx context.Context, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM applicant_works WHERE user_id = $1 AND kind = 'portrait'`, userID)
	if err != nil {
		return fmt.Errorf("delete applicant portrait: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
