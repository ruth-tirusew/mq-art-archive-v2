package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/outbound"
)

type OnboardingRepository struct {
	pool *Pool
}

func NewOnboardingRepository(pool *Pool) outbound.OnboardingRepository {
	return &OnboardingRepository{pool: pool}
}

const onboardingColumns = `
	id, applicant_id, applicant_type, display_name, COALESCE(requested_handle, ''), notes,
	status, reviewed_by, reviewed_at, created_at, updated_at
`

func (r *OnboardingRepository) ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+onboardingColumns+`
		FROM onboarding_applications
		WHERE status = $1
		ORDER BY created_at ASC
	`, string(onboarding.ApprovalStatusPending))
	if err != nil {
		return nil, fmt.Errorf("list pending applications: %w", err)
	}
	defer rows.Close()

	var apps []onboarding.OnboardingApplication
	for rows.Next() {
		app, err := scanOnboardingApplication(rows)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list pending applications: %w", err)
	}
	if apps == nil {
		apps = []onboarding.OnboardingApplication{}
	}
	return apps, nil
}

func (r *OnboardingRepository) GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+onboardingColumns+`
		FROM onboarding_applications
		WHERE id = $1
	`, id)

	app, err := scanOnboardingApplication(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get onboarding application: %w", err)
	}
	return &app, nil
}

func (r *OnboardingRepository) GetLatestByApplicantID(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+onboardingColumns+`
		FROM onboarding_applications
		WHERE applicant_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, applicantID)

	app, err := scanOnboardingApplication(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get latest onboarding application: %w", err)
	}
	return &app, nil
}

func (r *OnboardingRepository) Save(ctx context.Context, app onboarding.OnboardingApplication) (*onboarding.OnboardingApplication, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO onboarding_applications (
			id, applicant_id, applicant_type, display_name, requested_handle, notes,
			status, reviewed_by, reviewed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			requested_handle = EXCLUDED.requested_handle,
			notes = EXCLUDED.notes,
			status = EXCLUDED.status,
			reviewed_by = EXCLUDED.reviewed_by,
			reviewed_at = EXCLUDED.reviewed_at,
			updated_at = EXCLUDED.updated_at
		RETURNING `+onboardingColumns,
		app.ID, app.ApplicantID, string(app.ApplicantType), app.DisplayName, nullableString(app.RequestedHandle), app.Notes,
		string(app.Status), app.ReviewedBy, app.ReviewedAt, app.CreatedAt, app.UpdatedAt,
	)

	saved, err := scanOnboardingApplication(row)
	if err != nil {
		return nil, fmt.Errorf("save onboarding application: %w", err)
	}
	return &saved, nil
}

func scanOnboardingApplication(row scannable) (onboarding.OnboardingApplication, error) {
	var app onboarding.OnboardingApplication
	var applicantType, status string
	err := row.Scan(
		&app.ID, &app.ApplicantID, &applicantType, &app.DisplayName, &app.RequestedHandle, &app.Notes,
		&status, &app.ReviewedBy, &app.ReviewedAt, &app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return onboarding.OnboardingApplication{}, err
	}
	app.ApplicantType = onboarding.ApplicantType(applicantType)
	app.Status = onboarding.ApprovalStatus(status)
	return app, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
