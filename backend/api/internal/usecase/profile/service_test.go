package profile

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

type mockProfileRepo struct {
	getBySlug     func(ctx context.Context, slug string) (*domain.ArtistProfile, error)
	getByHandle   func(ctx context.Context, handle string) (*domain.ArtistProfile, error)
	listApproved  func(ctx context.Context, filter domain.ListFilter) ([]domain.ArtistProfile, error)
	getByID       func(ctx context.Context, id uuid.UUID) (*domain.ArtistProfile, error)
	getByUserID   func(ctx context.Context, userID uuid.UUID) (*domain.ArtistProfile, error)
	save          func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error)
}

func (m *mockProfileRepo) GetArtistBySlug(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
	return m.getBySlug(ctx, slug)
}

func (m *mockProfileRepo) GetArtistByHandle(ctx context.Context, handle string) (*domain.ArtistProfile, error) {
	if m.getByHandle != nil {
		return m.getByHandle(ctx, handle)
	}
	return nil, nil
}

func (m *mockProfileRepo) ListApproved(ctx context.Context, filter domain.ListFilter) ([]domain.ArtistProfile, error) {
	if m.listApproved != nil {
		return m.listApproved(ctx, filter)
	}
	return nil, nil
}

func (m *mockProfileRepo) GetArtistByID(ctx context.Context, id uuid.UUID) (*domain.ArtistProfile, error) {
	return m.getByID(ctx, id)
}

func (m *mockProfileRepo) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*domain.ArtistProfile, error) {
	return m.getByUserID(ctx, userID)
}

func (m *mockProfileRepo) SaveArtist(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
	return m.save(ctx, p)
}

func TestService_GetArtistBySlug_delegatesToRepository(t *testing.T) {
	expected := &domain.ArtistProfile{Slug: "abebe", DisplayName: "Abebe"}
	repo := &mockProfileRepo{
		getBySlug: func(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
			assist.Equal(t, "abebe", slug)
			return expected, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.GetArtistBySlug(context.Background(), "abebe")
	assist.NoError(t, err)
	assist.Equal(t, expected, got)
}

func TestService_UpdateArtist_delegatesToRepository(t *testing.T) {
	profile := domain.ArtistProfile{ID: uuid.New(), Slug: "abebe", DisplayName: "Abebe"}
	repo := &mockProfileRepo{
		save: func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
			assist.Equal(t, profile.ID, p.ID)
			return &p, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.UpdateArtist(context.Background(), profile)
	assist.NoError(t, err)
	assist.Equal(t, profile.ID, got.ID)
}
