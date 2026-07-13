package art

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/testutil/assist"
)

type mockArtPostRepo struct {
	listByArtist  func(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error)
	listPublished func(ctx context.Context, filter domain.ListFilter) ([]domain.ArtPostWithArtist, error)
	getByID       func(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error)
	create        func(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error)
}

func (m *mockArtPostRepo) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error) {
	if m.listByArtist != nil {
		return m.listByArtist(ctx, artistID)
	}
	return nil, nil
}

func (m *mockArtPostRepo) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.ArtPostWithArtist, error) {
	if m.listPublished != nil {
		return m.listPublished(ctx, filter)
	}
	return nil, nil
}

func (m *mockArtPostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error) {
	return m.getByID(ctx, id)
}

func (m *mockArtPostRepo) Create(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error) {
	return m.create(ctx, post)
}

func TestService_CreateDraft_requiresTitle(t *testing.T) {
	svc := NewService(&mockArtPostRepo{})

	_, err := svc.CreateDraft(context.Background(), uuid.New(), " ", "desc", "oil")
	assist.Error(t, err)
	assist.Contains(t, err.Error(), "title is required")
}

func TestService_CreateDraft_setsDraftStatus(t *testing.T) {
	artistID := uuid.New()
	var captured domain.ArtPost

	repo := &mockArtPostRepo{
		create: func(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error) {
			captured = post
			return &post, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.CreateDraft(context.Background(), artistID, "Sunset", "Over Addis", "watercolor")
	assist.NoError(t, err)

	assist.Equal(t, artistID, captured.ArtistID)
	assist.Equal(t, domain.ArtStatusDraft, captured.Status)
	assist.Equal(t, "Sunset", captured.Title)
}
