package profile

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

type mockProfileRepo struct {
	getBySlug    func(ctx context.Context, slug string) (*domain.ArtistProfile, error)
	getByHandle  func(ctx context.Context, handle string) (*domain.ArtistProfile, error)
	listApproved func(ctx context.Context, filter domain.ListFilter) ([]domain.ArtistProfile, error)
	getByID      func(ctx context.Context, id uuid.UUID) (*domain.ArtistProfile, error)
	getByUserID  func(ctx context.Context, userID uuid.UUID) (*domain.ArtistProfile, error)
	save         func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error)
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

func (m *mockProfileRepo) ListAll(ctx context.Context, status *domain.ProfileStatus, limit, offset int) ([]domain.ArtistProfile, error) {
	return nil, nil
}

func (m *mockProfileRepo) GetArtistByID(ctx context.Context, id uuid.UUID) (*domain.ArtistProfile, error) {
	if m.getByID == nil {
		return nil, apperrors.ErrNotFound
	}
	return m.getByID(ctx, id)
}

func (m *mockProfileRepo) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*domain.ArtistProfile, error) {
	return m.getByUserID(ctx, userID)
}

func (m *mockProfileRepo) SaveArtist(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
	return m.save(ctx, p)
}
func (m *mockProfileRepo) CreateDraftForUser(ctx context.Context, userID uuid.UUID, displayName string) (*domain.ArtistProfile, error) {
	return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, DisplayName: displayName, Slug: "draft"}, nil
}

func (m *mockProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestService_GetArtistBySlug_delegatesToRepository(t *testing.T) {
	expected := &domain.ArtistProfile{Slug: "abebe", DisplayName: "Abebe"}
	repo := &mockProfileRepo{
		getBySlug: func(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
			assist.Equal(t, "abebe", slug)
			return expected, nil
		},
	}
	svc := NewService(repo, nil, nil)

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
	svc := NewService(repo, nil, nil)

	got, err := svc.UpdateArtist(context.Background(), profile)
	assist.NoError(t, err)
	assist.Equal(t, profile.ID, got.ID)
}

func TestService_UpdateOwnProfile_publishRequiresBasics(t *testing.T) {
	userID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(ctx context.Context, uid uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
	}
	svc := NewService(repo, nil, nil)

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Status: domain.ProfileStatusApproved})
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestService_UpdateOwnProfile_publishesWhenComplete(t *testing.T) {
	userID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(ctx context.Context, uid uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
		getBySlug: func(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
		getByHandle: func(ctx context.Context, handle string) (*domain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
		save: func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
			assist.Equal(t, domain.ProfileStatusApproved, p.Status)
			assist.NotNil(t, p.ApprovedAt)
			return &p, nil
		},
	}
	svc := NewService(repo, nil, nil)

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{
		Bio:         "Painter",
		PortraitURL: "https://example.com/p.jpg",
		Status:      domain.ProfileStatusApproved,
	})
	assist.NoError(t, err)
}

func TestService_UpdateOwnProfile_rejectsPendingStatus(t *testing.T) {
	userID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(ctx context.Context, uid uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
	}
	svc := NewService(repo, nil, nil)

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Status: domain.ProfileStatusPending})
	assist.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestService_UpdateOwnProfile_detectsSlugConflict(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	otherID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(ctx context.Context, uid uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: profileID, UserID: userID, Slug: "artist", DisplayName: "Artist"}, nil
		},
		getBySlug: func(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
			if slug == "taken" {
				return &domain.ArtistProfile{ID: otherID, Slug: slug}, nil
			}
			return nil, apperrors.ErrNotFound
		},
		getByHandle: func(ctx context.Context, handle string) (*domain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	}
	svc := NewService(repo, nil, nil)

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Slug: "taken"})
	assist.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestService_UpdateOwnProfile_savesUpdates(t *testing.T) {
	userID := uuid.New()
	profileID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(ctx context.Context, uid uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: profileID, UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
		getBySlug: func(ctx context.Context, slug string) (*domain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
		getByHandle: func(ctx context.Context, handle string) (*domain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
		save: func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
			assist.Equal(t, "New Bio", p.Bio)
			assist.Equal(t, domain.ProfileStatusDraft, p.Status)
			return &p, nil
		},
	}
	svc := NewService(repo, nil, nil)

	got, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{
		Bio:    "New Bio",
		Status: domain.ProfileStatusDraft,
	})
	assist.NoError(t, err)
	assist.Equal(t, "New Bio", got.Bio)
}

func TestService_UpdateOwnProfile_cannotRepublishAProfileAnEditorTookDown(t *testing.T) {
	userID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(ctx context.Context, uid uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist",
				Status: domain.ProfileStatusDraft, ModerationHold: true}, nil
		},
		save: func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
			t.Fatal("a held profile must not be saved as published")
			return &p, nil
		},
	}
	svc := NewService(repo, nil, nil)

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{
		Bio: "Painter", PortraitURL: "https://example.com/p.jpg", Status: domain.ProfileStatusApproved,
	})
	assist.ErrorIs(t, err, apperrors.ErrForbidden)
}

func TestService_UpdateArtist_takedownSetsHoldAndApprovalReleasesIt(t *testing.T) {
	id := uuid.New()
	current := domain.ArtistProfile{ID: id, Status: domain.ProfileStatusApproved}
	var saved domain.ArtistProfile
	repo := &mockProfileRepo{
		getByID: func(ctx context.Context, _ uuid.UUID) (*domain.ArtistProfile, error) { c := current; return &c, nil },
		save: func(ctx context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
			saved = p
			return &p, nil
		},
	}
	svc := NewService(repo, nil, nil)

	_, err := svc.UpdateArtist(context.Background(), domain.ArtistProfile{ID: id, Status: domain.ProfileStatusDraft})
	assist.NoError(t, err)
	assist.Equal(t, true, saved.ModerationHold)

	current = saved // now taken down and held
	_, err = svc.UpdateArtist(context.Background(), domain.ArtistProfile{ID: id, Status: domain.ProfileStatusApproved})
	assist.NoError(t, err)
	assist.Equal(t, false, saved.ModerationHold)
}

// stubSanitizer stands in for the real one: it tags what it was given, so a test can tell the
// stored value came from the sanitizer, and reports `text` as the visible text.
type stubSanitizer struct{ text func(string) string }

func (s stubSanitizer) Sanitize(in string) (string, string) {
	if strings.TrimSpace(in) == "" || in == "<p></p>" {
		return "", ""
	}
	return "CLEAN(" + in + ")", s.text(in)
}

func TestService_UpdateOwnProfile_cleansTheBio(t *testing.T) {
	userID := uuid.New()
	var saved domain.ArtistProfile
	repo := &mockProfileRepo{
		getByUserID: func(context.Context, uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
		getBySlug:   func(context.Context, string) (*domain.ArtistProfile, error) { return nil, apperrors.ErrNotFound },
		getByHandle: func(context.Context, string) (*domain.ArtistProfile, error) { return nil, apperrors.ErrNotFound },
		save: func(_ context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) {
			saved = p
			return &p, nil
		},
	}
	svc := NewService(repo, nil, nil, WithRichText(stubSanitizer{text: func(s string) string { return s }}))

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Bio: "<p>hi</p><script>x</script>"})
	assist.NoError(t, err)
	assist.Equal(t, "CLEAN(<p>hi</p><script>x</script>)", saved.Bio) // what is stored is the sanitizer's output

	// A bio that is only empty markup is blank again, so it can't satisfy "add a bio before publishing".
	_, err = svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Bio: "<p></p>"})
	assist.NoError(t, err)
	assist.Equal(t, "", saved.Bio)
}

func TestService_UpdateOwnProfile_emptyMarkupDoesNotCountAsABioWhenPublishing(t *testing.T) {
	userID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(context.Context, uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
		getBySlug:   func(context.Context, string) (*domain.ArtistProfile, error) { return nil, apperrors.ErrNotFound },
		getByHandle: func(context.Context, string) (*domain.ArtistProfile, error) { return nil, apperrors.ErrNotFound },
	}
	svc := NewService(repo, nil, nil, WithRichText(stubSanitizer{text: func(s string) string { return s }}))

	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{
		Bio: "<p></p>", PortraitURL: "https://example.com/p.jpg", Status: domain.ProfileStatusApproved,
	})
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestService_UpdateOwnProfile_boundsTheVisibleBioLength(t *testing.T) {
	userID := uuid.New()
	repo := &mockProfileRepo{
		getByUserID: func(context.Context, uuid.UUID) (*domain.ArtistProfile, error) {
			return &domain.ArtistProfile{ID: uuid.New(), UserID: userID, Slug: "artist", DisplayName: "Artist", Status: domain.ProfileStatusDraft}, nil
		},
		getBySlug:   func(context.Context, string) (*domain.ArtistProfile, error) { return nil, apperrors.ErrNotFound },
		getByHandle: func(context.Context, string) (*domain.ArtistProfile, error) { return nil, apperrors.ErrNotFound },
		save:        func(_ context.Context, p domain.ArtistProfile) (*domain.ArtistProfile, error) { return &p, nil },
	}
	// Amharic letters are 3 bytes each: the limit counts characters, not bytes.
	svc := NewService(repo, nil, nil, WithRichText(stubSanitizer{text: func(string) string { return strings.Repeat("ሰ", maxBioText) }}))
	_, err := svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Bio: "<p>x</p>"})
	assist.NoError(t, err)

	svc = NewService(repo, nil, nil, WithRichText(stubSanitizer{text: func(string) string { return strings.Repeat("ሰ", maxBioText+1) }}))
	_, err = svc.UpdateOwnProfile(context.Background(), userID, domain.OwnProfileUpdate{Bio: "<p>x</p>"})
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}
