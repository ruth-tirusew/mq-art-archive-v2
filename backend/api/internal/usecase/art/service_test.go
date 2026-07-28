package art

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/testutil/assist"
)

type mockArtPostRepo struct {
	listByArtist      func(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error)
	listOwnedByArtist func(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error)
	listPublished     func(ctx context.Context, filter domain.ListFilter) ([]domain.ArtPostWithArtist, error)
	listAll           func(ctx context.Context, status *domain.ArtStatus, limit, offset int) ([]domain.ArtPostWithArtist, error)
	getByID           func(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error)
	create            func(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error)
	update            func(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error)
	delete            func(ctx context.Context, id uuid.UUID) error
}

func (m *mockArtPostRepo) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error) {
	if m.listByArtist != nil {
		return m.listByArtist(ctx, artistID)
	}
	return nil, nil
}

func (m *mockArtPostRepo) ListOwnedByArtist(ctx context.Context, artistID uuid.UUID) ([]domain.ArtPost, error) {
	if m.listOwnedByArtist != nil {
		return m.listOwnedByArtist(ctx, artistID)
	}
	return nil, nil
}

func (m *mockArtPostRepo) ListPublished(ctx context.Context, filter domain.ListFilter) ([]domain.ArtPostWithArtist, error) {
	if m.listPublished != nil {
		return m.listPublished(ctx, filter)
	}
	return nil, nil
}

func (m *mockArtPostRepo) ListAll(ctx context.Context, status *domain.ArtStatus, limit, offset int) ([]domain.ArtPostWithArtist, error) {
	if m.listAll != nil {
		return m.listAll(ctx, status, limit, offset)
	}
	return nil, nil
}

func (m *mockArtPostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error) {
	if m.getByID != nil {
		return m.getByID(ctx, id)
	}
	return nil, nil
}

func (m *mockArtPostRepo) Create(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error) {
	if m.create != nil {
		return m.create(ctx, post)
	}
	return &post, nil
}

func (m *mockArtPostRepo) Update(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error) {
	if m.update != nil {
		return m.update(ctx, post)
	}
	return &post, nil
}

func (m *mockArtPostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.delete != nil {
		return m.delete(ctx, id)
	}
	return nil
}

func TestService_CreateDraft_requiresTitle(t *testing.T) {
	svc := NewService(&mockArtPostRepo{})

	_, err := svc.CreateDraft(context.Background(), uuid.New(), domain.ArtPostWrite{Title: " ", Description: "desc", Medium: "oil"})
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

	_, err := svc.CreateDraft(context.Background(), artistID, domain.ArtPostWrite{Title: "Sunset", Description: "Over Addis", Medium: "watercolor"})
	assist.NoError(t, err)

	assist.Equal(t, artistID, captured.ArtistID)
	assist.Equal(t, domain.ArtStatusDraft, captured.Status)
	assist.Equal(t, "Sunset", captured.Title)
}

func TestService_AdminCreate_published(t *testing.T) {
	artistID := uuid.New()
	var captured domain.ArtPost
	repo := &mockArtPostRepo{
		create: func(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error) {
			captured = post
			return &post, nil
		},
	}
	svc := NewService(repo)
	status := domain.ArtStatusPublished
	_, err := svc.AdminCreate(context.Background(), artistID, domain.ArtPostWrite{Title: "Market"}, &status)
	assist.NoError(t, err)
	assist.Equal(t, domain.ArtStatusPublished, captured.Status)
	assist.NotNil(t, captured.PublishedAt)
}

func TestService_AdminDelete(t *testing.T) {
	postID := uuid.New()
	deleted := false
	repo := &mockArtPostRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error) {
			return &domain.ArtPost{ID: id, Title: "X"}, nil
		},
		delete: func(ctx context.Context, id uuid.UUID) error {
			assist.Equal(t, postID, id)
			deleted = true
			return nil
		},
	}
	svc := NewService(repo)
	assist.NoError(t, svc.AdminDelete(context.Background(), postID))
	assist.Equal(t, true, deleted)
}
