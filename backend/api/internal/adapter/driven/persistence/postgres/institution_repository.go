package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/port/outbound"
)

type InstitutionRepository struct {
	pool *Pool
}

func NewInstitutionRepository(pool *Pool) outbound.InstitutionRepository {
	return &InstitutionRepository{pool: pool}
}

const institutionColumns = `
	id, user_id, slug, name, description,
	contact_email, contact_phone, contact_website, contact_location,
	status, created_at, updated_at
`

func (r *InstitutionRepository) GetBySlug(ctx context.Context, slug string) (*institution.Institution, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+institutionColumns+`
		FROM institution_profiles
		WHERE slug = $1
	`, slug)

	inst, err := scanInstitution(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get institution by slug: %w", err)
	}
	return &inst, nil
}

func (r *InstitutionRepository) GetByID(ctx context.Context, id uuid.UUID) (*institution.Institution, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+institutionColumns+`
		FROM institution_profiles
		WHERE id = $1
	`, id)

	inst, err := scanInstitution(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get institution by id: %w", err)
	}
	return &inst, nil
}

func (r *InstitutionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*institution.Institution, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+institutionColumns+`
		FROM institution_profiles
		WHERE user_id = $1
	`, userID)

	inst, err := scanInstitution(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get institution by user id: %w", err)
	}
	return &inst, nil
}

func (r *InstitutionRepository) Save(ctx context.Context, inst institution.Institution) (*institution.Institution, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO institution_profiles (
			id, user_id, slug, name, description,
			contact_email, contact_phone, contact_website, contact_location,
			status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			contact_email = EXCLUDED.contact_email,
			contact_phone = EXCLUDED.contact_phone,
			contact_website = EXCLUDED.contact_website,
			contact_location = EXCLUDED.contact_location,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
		RETURNING `+institutionColumns,
		inst.ID, inst.UserID, inst.Slug, inst.Name, inst.Description,
		inst.Contact.Email, inst.Contact.Phone, inst.Contact.Website, inst.Contact.Location,
		string(inst.Status), inst.CreatedAt, inst.UpdatedAt,
	)

	saved, err := scanInstitution(row)
	if err != nil {
		return nil, fmt.Errorf("save institution: %w", err)
	}
	return &saved, nil
}

func (r *InstitutionRepository) CreateDraftForUser(ctx context.Context, userID uuid.UUID, displayName string) (*institution.Institution, error) {
	now := time.Now().UTC()
	slug := slugify(displayName)
	return r.Save(ctx, institution.Institution{
		ID:        uuid.New(),
		UserID:    userID,
		Slug:      slug,
		Name:      displayName,
		Status:    institution.InstitutionStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func scanInstitution(row scannable) (institution.Institution, error) {
	var inst institution.Institution
	var status string
	err := row.Scan(
		&inst.ID, &inst.UserID, &inst.Slug, &inst.Name, &inst.Description,
		&inst.Contact.Email, &inst.Contact.Phone, &inst.Contact.Website, &inst.Contact.Location,
		&status, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		return institution.Institution{}, err
	}
	inst.Status = institution.InstitutionStatus(status)
	return inst, nil
}
