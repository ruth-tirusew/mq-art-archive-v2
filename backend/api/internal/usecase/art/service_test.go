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

func TestMediaFromURLs_preservesExistingIDsByURL(t *testing.T) {
	keptID := uuid.New()
	removedURL := "https://example.com/removed.jpg"
	existing := []domain.MediaAsset{
		{ID: keptID, URL: "https://example.com/kept.jpg", SortOrder: 0},
		{ID: uuid.New(), URL: removedURL, SortOrder: 1},
	}

	out := mediaFromURLs([]string{"https://example.com/kept.jpg", "https://example.com/new.jpg"}, existing)

	assist.Len(t, 2, len(out))
	assist.Equal(t, keptID, out[0].ID)
	assist.Equal(t, "https://example.com/kept.jpg", out[0].URL)
	assist.NotEqual(t, uuid.Nil, out[1].ID)
	assist.Equal(t, "https://example.com/new.jpg", out[1].URL)
	for _, m := range out {
		if m.URL == removedURL {
			t.Fatalf("removed URL should not appear in the result: %v", out)
		}
	}
}

func TestMediaFromURLs_duplicateURLsConsumeDistinctExistingIDs(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	existing := []domain.MediaAsset{
		{ID: firstID, URL: "https://example.com/dup.jpg", SortOrder: 0},
		{ID: secondID, URL: "https://example.com/dup.jpg", SortOrder: 1},
	}

	out := mediaFromURLs([]string{"https://example.com/dup.jpg", "https://example.com/dup.jpg"}, existing)

	assist.Len(t, 2, len(out))
	assist.Equal(t, firstID, out[0].ID)
	assist.Equal(t, secondID, out[1].ID)
}

func TestService_UpdateOwned_preservesMediaIDForUnchangedURL(t *testing.T) {
	artistID := uuid.New()
	postID := uuid.New()
	keptID := uuid.New()
	const keptURL = "https://example.com/kept.jpg"

	var captured domain.ArtPost
	repo := &mockArtPostRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.ArtPost, error) {
			return &domain.ArtPost{
				ID:       id,
				ArtistID: artistID,
				Title:    "Original",
				Media:    []domain.MediaAsset{{ID: keptID, URL: keptURL, SortOrder: 0}},
			}, nil
		},
		update: func(ctx context.Context, post domain.ArtPost) (*domain.ArtPost, error) {
			captured = post
			return &post, nil
		},
	}
	svc := NewService(repo)

	_, err := svc.UpdateOwned(context.Background(), artistID, postID, domain.ArtPostWrite{
		Title:     "Updated",
		MediaURLs: []string{keptURL, "https://example.com/new.jpg"},
	})
	assist.NoError(t, err)

	assist.Len(t, 2, len(captured.Media))
	assist.Equal(t, keptID, captured.Media[0].ID)
	assist.NotEqual(t, uuid.Nil, captured.Media[1].ID)
	assist.NotEqual(t, keptID, captured.Media[1].ID)
}
