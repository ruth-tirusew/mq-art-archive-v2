package institution

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/testutil/assist"
)

type mockInstitutionRepo struct {
	getBySlug          func(ctx context.Context, slug string) (*domain.Institution, error)
	getByID            func(ctx context.Context, id uuid.UUID) (*domain.Institution, error)
	getByUserID        func(ctx context.Context, userID uuid.UUID) (*domain.Institution, error)
	save               func(ctx context.Context, inst domain.Institution) (*domain.Institution, error)
	createDraftForUser func(ctx context.Context, userID uuid.UUID, displayName string) (*domain.Institution, error)
}

func (m *mockInstitutionRepo) GetBySlug(ctx context.Context, slug string) (*domain.Institution, error) {
	return m.getBySlug(ctx, slug)
}

func (m *mockInstitutionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Institution, error) {
	return m.getByID(ctx, id)
}

func (m *mockInstitutionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Institution, error) {
	if m.getByUserID != nil {
		return m.getByUserID(ctx, userID)
	}
	return nil, nil
}

func (m *mockInstitutionRepo) Save(ctx context.Context, inst domain.Institution) (*domain.Institution, error) {
	return m.save(ctx, inst)
}

func (m *mockInstitutionRepo) CreateDraftForUser(ctx context.Context, userID uuid.UUID, displayName string) (*domain.Institution, error) {
	if m.createDraftForUser != nil {
		return m.createDraftForUser(ctx, userID, displayName)
	}
	return &domain.Institution{UserID: userID, Name: displayName}, nil
}

func TestService_GetBySlug_delegatesToRepository(t *testing.T) {
	expected := &domain.Institution{Slug: "national-gallery", Name: "National Gallery"}
	repo := &mockInstitutionRepo{
		getBySlug: func(ctx context.Context, slug string) (*domain.Institution, error) {
			assist.Equal(t, "national-gallery", slug)
			return expected, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.GetBySlug(context.Background(), "national-gallery")
	assist.NoError(t, err)
	assist.Equal(t, expected, got)
}

func TestService_Update_delegatesToRepository(t *testing.T) {
	inst := domain.Institution{ID: uuid.New(), Slug: "studio", Name: "Studio"}
	repo := &mockInstitutionRepo{
		save: func(ctx context.Context, i domain.Institution) (*domain.Institution, error) {
			assist.Equal(t, inst.ID, i.ID)
			return &i, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.Update(context.Background(), inst)
	assist.NoError(t, err)
	assist.Equal(t, inst.ID, got.ID)
}
