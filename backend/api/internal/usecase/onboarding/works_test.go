package onboarding

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/onboarding"
	profiledomain "github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/testutil/assist"
)

type fakeWorks struct {
	byUser map[uuid.UUID][]domain.ApplicantWork
	// owned is the set of media assets each user may use; anything else is "not found".
	owned     map[uuid.UUID]uuid.UUID
	portraits map[uuid.UUID]domain.ApplicantPortrait
}

func newFakeWorks() *fakeWorks {
	return &fakeWorks{
		byUser: map[uuid.UUID][]domain.ApplicantWork{}, owned: map[uuid.UUID]uuid.UUID{},
		portraits: map[uuid.UUID]domain.ApplicantPortrait{},
	}
}

func (f *fakeWorks) GetPortrait(_ context.Context, u uuid.UUID) (*domain.ApplicantPortrait, error) {
	if p, ok := f.portraits[u]; ok {
		return &p, nil
	}
	return nil, apperrors.ErrNotFound
}

func (f *fakeWorks) SetPortrait(_ context.Context, u, media uuid.UUID) (*domain.ApplicantPortrait, error) {
	if f.owned[media] != u {
		return nil, apperrors.ErrNotFound
	}
	p := domain.ApplicantPortrait{MediaAssetID: media, URL: "https://res.cloudinary.com/c/image/upload/" + media.String()}
	f.portraits[u] = p
	return &p, nil
}

func (f *fakeWorks) DeletePortrait(_ context.Context, u uuid.UUID) error {
	if _, ok := f.portraits[u]; !ok {
		return apperrors.ErrNotFound
	}
	delete(f.portraits, u)
	return nil
}

func (f *fakeWorks) ListByUser(_ context.Context, u uuid.UUID) ([]domain.ApplicantWork, error) {
	return append([]domain.ApplicantWork{}, f.byUser[u]...), nil
}

func (f *fakeWorks) Add(_ context.Context, w domain.ApplicantWork) (*domain.ApplicantWork, error) {
	if f.owned[w.MediaAssetID] != w.UserID {
		return nil, apperrors.ErrNotFound
	}
	for _, e := range f.byUser[w.UserID] {
		if e.MediaAssetID == w.MediaAssetID {
			return nil, apperrors.ErrConflict
		}
	}
	w.Position = len(f.byUser[w.UserID])
	w.URL = "https://res.cloudinary.com/c/image/upload/" + w.MediaAssetID.String()
	f.byUser[w.UserID] = append(f.byUser[w.UserID], w)
	return &w, nil
}

func (f *fakeWorks) Delete(_ context.Context, u, id uuid.UUID) error {
	for i, w := range f.byUser[u] {
		if w.ID == id {
			f.byUser[u] = append(f.byUser[u][:i], f.byUser[u][i+1:]...)
			return nil
		}
	}
	return apperrors.ErrNotFound
}

// upload registers an image the user owns and returns its id.
func (f *fakeWorks) upload(user uuid.UUID) uuid.UUID {
	id := uuid.New()
	f.owned[id] = user
	return id
}

type worksEnv struct {
	svc    *Service
	works  *fakeWorks
	repo   *mockOnboardingRepo
	saved  []domain.OnboardingApplication
	latest *domain.OnboardingApplication
}

func newWorksEnv(t *testing.T) *worksEnv {
	t.Helper()
	e := &worksEnv{works: newFakeWorks()}
	e.repo = &mockOnboardingRepo{
		getLatestByApplicantID: func(context.Context, uuid.UUID) (*domain.OnboardingApplication, error) {
			if e.latest == nil {
				return nil, apperrors.ErrNotFound
			}
			return e.latest, nil
		},
		save: func(_ context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			e.saved = append(e.saved, app)
			return &app, nil
		},
	}
	e.svc = NewService(e.repo, &mockUserRepo{}, &mockProfileRepo{
		getByUserID: func(context.Context, uuid.UUID) (*profiledomain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	}, &mockInstitutionRepo{}, WithApplicantWorks(e.works))
	return e
}

func (e *worksEnv) addPortrait(t *testing.T, user uuid.UUID) {
	t.Helper()
	_, err := e.svc.SetPortrait(context.Background(), user, e.works.upload(user))
	assist.NoError(t, err)
}

func (e *worksEnv) addWorks(t *testing.T, user uuid.UUID, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		_, err := e.svc.AddWork(context.Background(), user, inbound.AddWorkInput{MediaAssetID: e.works.upload(user), Title: "Work"})
		assist.NoError(t, err)
	}
}

func TestSubmit_artistNeedsFiveWorks(t *testing.T) {
	e := newWorksEnv(t)
	user := uuid.New()
	e.addWorks(t, user, 4)

	_, err := e.svc.SubmitWithHandle(context.Background(), user, domain.ApplicantTypeArtist, "Ada", "ada_1", "")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Contains(t, err.Error(), "at least 5")
	assist.Contains(t, err.Error(), "you have 4")
	assist.Len(t, 0, len(e.saved)) // nothing was saved

	e.addWorks(t, user, 1)
	e.addPortrait(t, user)
	app, err := e.svc.SubmitWithHandle(context.Background(), user, domain.ApplicantTypeArtist, "Ada", "ada_1", "")
	assist.NoError(t, err)
	assist.Len(t, 5, len(app.Works)) // frozen onto the application
}

func TestSubmit_artistNeedsAPortrait(t *testing.T) {
	e := newWorksEnv(t)
	user := uuid.New()
	e.addWorks(t, user, 5)

	_, err := e.svc.SubmitWithHandle(context.Background(), user, domain.ApplicantTypeArtist, "Ada", "ada_1", "")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Contains(t, err.Error(), "portrait")
	assist.Len(t, 0, len(e.saved))

	e.addPortrait(t, user)
	app, err := e.svc.SubmitWithHandle(context.Background(), user, domain.ApplicantTypeArtist, "Ada", "ada_1", "")
	assist.NoError(t, err)
	if app.Portrait == nil || app.Portrait.URL == "" {
		t.Fatalf("the portrait must be frozen onto the application, got %+v", app.Portrait)
	}
}

func TestPortrait_isNotOneOfTheFiveWorks(t *testing.T) {
	e := newWorksEnv(t)
	user := uuid.New()
	e.addWorks(t, user, 4)
	e.addPortrait(t, user)

	// Four works and a portrait is still four works.
	_, err := e.svc.SubmitWithHandle(context.Background(), user, domain.ApplicantTypeArtist, "Ada", "ada_1", "")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Contains(t, err.Error(), "at least 5")
}

func TestPortrait_setReplaceRemoveAndLock(t *testing.T) {
	e := newWorksEnv(t)
	ctx := context.Background()
	user := uuid.New()

	_, err := e.svc.SetPortrait(ctx, user, e.works.upload(uuid.New())) // someone else's image
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
	_, err = e.svc.SetPortrait(ctx, user, uuid.Nil)
	assist.ErrorIs(t, err, apperrors.ErrValidation)

	first := e.works.upload(user)
	_, err = e.svc.SetPortrait(ctx, user, first)
	assist.NoError(t, err)
	second := e.works.upload(user)
	got, err := e.svc.SetPortrait(ctx, user, second)
	assist.NoError(t, err)
	assist.Equal(t, second, got.MediaAssetID) // replaced, not added

	// While the application is being reviewed, what the reviewer sees can't change.
	e.latest = &domain.OnboardingApplication{Status: domain.ApprovalStatusPending}
	_, err = e.svc.SetPortrait(ctx, user, e.works.upload(user))
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	assist.ErrorIs(t, e.svc.RemovePortrait(ctx, user), apperrors.ErrConflict)

	e.latest = nil
	assist.NoError(t, e.svc.RemovePortrait(ctx, user))
	_, err = e.svc.GetPortrait(ctx, user)
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestSubmit_freezesTheWorksSoLaterChangesDontRewriteTheApplication(t *testing.T) {
	e := newWorksEnv(t)
	user := uuid.New()
	e.addWorks(t, user, 5)
	e.addPortrait(t, user)
	app, err := e.svc.SubmitWithHandle(context.Background(), user, domain.ApplicantTypeArtist, "Ada", "ada_1", "")
	assist.NoError(t, err)

	// After a rejection the applicant can change their works; the old application keeps its own.
	e.latest = &domain.OnboardingApplication{Status: domain.ApprovalStatusRejected}
	assist.NoError(t, e.svc.RemoveWork(context.Background(), user, e.works.byUser[user][0].ID))
	assist.Len(t, 4, len(e.works.byUser[user]))
	assist.Len(t, 5, len(app.Works))
}

func TestSubmit_institutionsAndUnconfiguredServicesAreNotGated(t *testing.T) {
	e := newWorksEnv(t)
	_, err := e.svc.SubmitWithHandle(context.Background(), uuid.New(), domain.ApplicantTypeInstitution, "Gallery", "", "")
	assist.NoError(t, err)

	plain := NewService(e.repo, &mockUserRepo{}, &mockProfileRepo{
		getByUserID: func(context.Context, uuid.UUID) (*profiledomain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	}, &mockInstitutionRepo{})
	_, err = plain.SubmitWithHandle(context.Background(), uuid.New(), domain.ApplicantTypeArtist, "Ada", "ada_2", "")
	assist.NoError(t, err)
}

func TestAddWork_validatesAndOnlyAcceptsTheUsersOwnImages(t *testing.T) {
	e := newWorksEnv(t)
	user, other := uuid.New(), uuid.New()
	ctx := context.Background()

	_, err := e.svc.AddWork(ctx, user, inbound.AddWorkInput{MediaAssetID: e.works.upload(user), Title: "  "})
	assist.ErrorIs(t, err, apperrors.ErrValidation)

	_, err = e.svc.AddWork(ctx, user, inbound.AddWorkInput{MediaAssetID: e.works.upload(other), Title: "Not mine"})
	assist.ErrorIs(t, err, apperrors.ErrNotFound)

	asset := e.works.upload(user)
	got, err := e.svc.AddWork(ctx, user, inbound.AddWorkInput{MediaAssetID: asset, Title: "  Untitled  ", Medium: " oil "})
	assist.NoError(t, err)
	assist.Equal(t, "Untitled", got.Title)
	assist.Equal(t, "oil", got.Medium)

	_, err = e.svc.AddWork(ctx, user, inbound.AddWorkInput{MediaAssetID: asset, Title: "Again"})
	assist.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestAddWork_stopsAtTheLimit(t *testing.T) {
	e := newWorksEnv(t)
	user := uuid.New()
	e.addWorks(t, user, domain.MaxApplicationWorks)
	_, err := e.svc.AddWork(context.Background(), user, inbound.AddWorkInput{MediaAssetID: e.works.upload(user), Title: "One too many"})
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Contains(t, err.Error(), "up to 20")
}

func TestWorks_areLockedWhileTheApplicationIsUnderReview(t *testing.T) {
	e := newWorksEnv(t)
	user := uuid.New()
	e.addWorks(t, user, 5)
	e.latest = &domain.OnboardingApplication{Status: domain.ApprovalStatusPending}

	_, err := e.svc.AddWork(context.Background(), user, inbound.AddWorkInput{MediaAssetID: e.works.upload(user), Title: "Late"})
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	assist.Contains(t, err.Error(), "being reviewed")
	assist.ErrorIs(t, e.svc.RemoveWork(context.Background(), user, e.works.byUser[user][0].ID), apperrors.ErrConflict)
	assist.Len(t, 5, len(e.works.byUser[user]))

	// Listing is always allowed.
	got, err := e.svc.ListWorks(context.Background(), user)
	assist.NoError(t, err)
	assist.Len(t, 5, len(got))
}

func TestRemoveWork_cannotTouchSomeoneElsesWork(t *testing.T) {
	e := newWorksEnv(t)
	owner, intruder := uuid.New(), uuid.New()
	e.addWorks(t, owner, 1)
	err := e.svc.RemoveWork(context.Background(), intruder, e.works.byUser[owner][0].ID)
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
	assist.Len(t, 1, len(e.works.byUser[owner]))
}

func TestWorks_notImplementedWithoutARepository(t *testing.T) {
	svc := NewService(&mockOnboardingRepo{}, &mockUserRepo{}, &mockProfileRepo{}, &mockInstitutionRepo{})
	_, err := svc.ListWorks(context.Background(), uuid.New())
	assist.ErrorIs(t, err, apperrors.ErrNotImplemented)
	assert := strings.Contains(err.Error(), "not implemented")
	if !assert {
		t.Fatal("expected the standard not-implemented error")
	}
}
