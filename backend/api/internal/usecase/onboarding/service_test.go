package onboarding

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/testutil/assist"
)

type mockOnboardingRepo struct {
	listPending func(ctx context.Context) ([]domain.OnboardingApplication, error)
	getByID     func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error)
	save        func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error)
}

func (m *mockOnboardingRepo) ListPending(ctx context.Context) ([]domain.OnboardingApplication, error) {
	return m.listPending(ctx)
}

func (m *mockOnboardingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	return m.getByID(ctx, id)
}

func (m *mockOnboardingRepo) Save(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
	return m.save(ctx, app)
}

func TestService_Review_updatesApplication(t *testing.T) {
	appID := uuid.New()
	reviewerID := uuid.New()
	existing := &domain.OnboardingApplication{
		ID:          appID,
		DisplayName: "Studio X",
		Status:      domain.ApprovalStatusPending,
	}

	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return existing, nil
		},
		save: func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			assist.Equal(t, domain.ApprovalStatusApproved, app.Status)
			assist.Equal(t, "looks good", app.Notes)
			assist.Equal(t, reviewerID, *app.ReviewedBy)
			assist.NotNil(t, app.ReviewedAt)
			return &app, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.Review(context.Background(), appID, reviewerID, domain.ApprovalStatusApproved, "looks good")
	assist.NoError(t, err)
	assist.Equal(t, domain.ApprovalStatusApproved, got.Status)
	assist.WithinDuration(t, time.Now().UTC(), got.UpdatedAt, time.Second)
}
