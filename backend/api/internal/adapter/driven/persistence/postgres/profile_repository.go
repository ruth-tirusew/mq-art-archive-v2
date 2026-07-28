package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/outbound"
)

type ProfileRepository struct {
	pool *Pool
}

func NewProfileRepository(pool *Pool) outbound.ProfileRepository {
	return &ProfileRepository{pool: pool}
}

const artistProfileColumns = `
	id, user_id, slug, COALESCE(handle, '') AS handle, display_name, COALESCE(bio, '') AS bio, born, discipline, tagline, years_active, COALESCE(portrait_url, '') AS portrait_url, featured,
	COALESCE(influences, '{}') AS influences, in_residence, COALESCE(residency_place, '') AS residency_place, open_for_commission,
	COALESCE(contact_email, '') AS contact_email, COALESCE(contact_phone, '') AS contact_phone, COALESCE(contact_website, '') AS contact_website, COALESCE(contact_location, '') AS contact_location,
	COALESCE(social_instagram, '') AS social_instagram, COALESCE(social_twitter, '') AS social_twitter, COALESCE(social_telegram, '') AS social_telegram,
	status, approved_at, created_at, updated_at
`

func (r *ProfileRepository) GetArtistByHandle(ctx context.Context, handle string) (*profile.ArtistProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+artistProfileColumns+`
		FROM artist_profiles
		WHERE handle = $1 OR slug = $1
	`, handle)

	p, err := scanArtistProfile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get artist by handle: %w", err)
	}
	return &p, nil
}

func (r *ProfileRepository) ListApproved(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 24
	}
	if limit > 50 {
		limit = 50
	}
	query := `
		SELECT ` + artistProfileColumns + `
		FROM artist_profiles
		WHERE status = $1`
	args := []any{string(profile.ProfileStatusApproved)}
	argPos := 2
	if filter.Featured != nil && *filter.Featured {
		query += fmt.Sprintf(" AND featured = $%d", argPos)
		args = append(args, true)
		argPos++
	}
	if filter.Query != "" {
		query += fmt.Sprintf(" AND (display_name ILIKE $%d OR bio ILIKE $%d OR discipline ILIKE $%d OR slug ILIKE $%d OR handle ILIKE $%d)", argPos, argPos, argPos, argPos, argPos)
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}
	query += " ORDER BY display_name ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list approved artists: %w", err)
	}
	defer rows.Close()

	var profiles []profile.ArtistProfile
	for rows.Next() {
		p, err := scanArtistProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	if profiles == nil {
		profiles = []profile.ArtistProfile{}
	}
	return profiles, rows.Err()
}

func (r *ProfileRepository) CountApproved(ctx context.Context, filter profile.ListFilter) (int, error) {
	query := `SELECT COUNT(*) FROM artist_profiles WHERE status = $1`
	args := []any{string(profile.ProfileStatusApproved)}
	argPos := 2
	if filter.Featured != nil && *filter.Featured {
		query += fmt.Sprintf(" AND featured = $%d", argPos)
		args = append(args, true)
		argPos++
	}
	if filter.Query != "" {
		query += fmt.Sprintf(" AND (display_name ILIKE $%d OR bio ILIKE $%d OR discipline ILIKE $%d OR slug ILIKE $%d OR handle ILIKE $%d)", argPos, argPos, argPos, argPos, argPos)
		args = append(args, "%"+filter.Query+"%")
	}
	var total int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count approved artists: %w", err)
	}
	return total, nil
}

func (r *ProfileRepository) ListAll(ctx context.Context, status *profile.ProfileStatus, limit, offset int) ([]profile.ArtistProfile, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT ` + artistProfileColumns + `
		FROM artist_profiles
		WHERE 1=1`
	args := []any{}
	argPos := 1
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, string(*status))
		argPos++
	}
	query += " ORDER BY updated_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all artists: %w", err)
	}
	defer rows.Close()

	var profiles []profile.ArtistProfile
	for rows.Next() {
		p, err := scanArtistProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	if profiles == nil {
		profiles = []profile.ArtistProfile{}
	}
	return profiles, rows.Err()
}

func (r *ProfileRepository) GetArtistBySlug(ctx context.Context, slug string) (*profile.ArtistProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+artistProfileColumns+`
		FROM artist_profiles
		WHERE slug = $1
	`, slug)

	p, err := scanArtistProfile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get artist by slug: %w", err)
	}
	return &p, nil
}

func (r *ProfileRepository) GetArtistByID(ctx context.Context, id uuid.UUID) (*profile.ArtistProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+artistProfileColumns+`
		FROM artist_profiles
		WHERE id = $1
	`, id)

	p, err := scanArtistProfile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get artist by id: %w", err)
	}
	return &p, nil
}

func (r *ProfileRepository) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profile.ArtistProfile, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+artistProfileColumns+`
		FROM artist_profiles
		WHERE user_id = $1
	`, userID)

	p, err := scanArtistProfile(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get artist by user id: %w", err)
	}
	return &p, nil
}

func (r *ProfileRepository) SaveArtist(ctx context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error) {
	influences := p.Influences
	if influences == nil {
		influences = []string{}
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO artist_profiles (
			id, user_id, slug, handle, display_name, bio, born, discipline, tagline, years_active, portrait_url, featured,
			influences, in_residence, residency_place, open_for_commission,
			contact_email, contact_phone, contact_website, contact_location,
			social_instagram, social_twitter, social_telegram,
			status, approved_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			handle = EXCLUDED.handle,
			display_name = EXCLUDED.display_name,
			bio = EXCLUDED.bio,
			born = EXCLUDED.born,
			discipline = EXCLUDED.discipline,
			tagline = EXCLUDED.tagline,
			years_active = EXCLUDED.years_active,
			portrait_url = EXCLUDED.portrait_url,
			featured = EXCLUDED.featured,
			influences = EXCLUDED.influences,
			in_residence = EXCLUDED.in_residence,
			residency_place = EXCLUDED.residency_place,
			open_for_commission = EXCLUDED.open_for_commission,
			contact_email = EXCLUDED.contact_email,
			contact_phone = EXCLUDED.contact_phone,
			contact_website = EXCLUDED.contact_website,
			contact_location = EXCLUDED.contact_location,
			social_instagram = EXCLUDED.social_instagram,
			social_twitter = EXCLUDED.social_twitter,
			social_telegram = EXCLUDED.social_telegram,
			status = EXCLUDED.status,
			approved_at = EXCLUDED.approved_at,
			updated_at = EXCLUDED.updated_at
		RETURNING `+artistProfileColumns,
		p.ID, p.UserID, p.Slug, nullIfEmpty(p.Handle, p.Slug), p.DisplayName, p.Bio, p.Born, p.Discipline, p.Tagline, p.YearsActive, p.PortraitURL, p.Featured,
		influences, p.InResidence, p.ResidencyPlace, p.OpenForCommission,
		p.Contact.Email, p.Contact.Phone, p.Contact.Website, p.Contact.Location,
		p.Social.Instagram, p.Social.Twitter, p.Social.Telegram,
		string(p.Status), p.ApprovedAt, p.CreatedAt, p.UpdatedAt,
	)

	saved, err := scanArtistProfile(row)
	if err != nil {
		return nil, fmt.Errorf("save artist profile: %w", err)
	}
	return &saved, nil
}

func (r *ProfileRepository) CreateDraftForUser(ctx context.Context, userID uuid.UUID, displayName string) (*profile.ArtistProfile, error) {
	now := time.Now().UTC()
	slug := slugify(displayName)
	return r.SaveArtist(ctx, profile.ArtistProfile{
		ID:          uuid.New(),
		UserID:      userID,
		Slug:        slug,
		Handle:      slug,
		DisplayName: displayName,
		Status:      profile.ProfileStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (r *ProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM artist_profiles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete artist profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanArtistProfile(row scannable) (profile.ArtistProfile, error) {
	var p profile.ArtistProfile
	var status string
	err := row.Scan(
		&p.ID, &p.UserID, &p.Slug, &p.Handle, &p.DisplayName, &p.Bio, &p.Born, &p.Discipline, &p.Tagline, &p.YearsActive, &p.PortraitURL, &p.Featured,
		&p.Influences, &p.InResidence, &p.ResidencyPlace, &p.OpenForCommission,
		&p.Contact.Email, &p.Contact.Phone, &p.Contact.Website, &p.Contact.Location,
		&p.Social.Instagram, &p.Social.Twitter, &p.Social.Telegram,
		&status, &p.ApprovedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return profile.ArtistProfile{}, err
	}
	if p.Influences == nil {
		p.Influences = []string{}
	}
	p.Status = profile.ProfileStatus(status)
	return p, nil
}

func nullIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func slugify(value string) string {
	s := strings.ToLower(strings.TrimSpace(value))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
