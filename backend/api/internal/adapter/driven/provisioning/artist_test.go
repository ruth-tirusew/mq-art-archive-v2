package provisioning

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/outbound"
)

type fakeProfiles struct {
	outbound.ProfileRepository
	byUser map[uuid.UUID]*profile.ArtistProfile
	saves  int
}

func (f *fakeProfiles) GetArtistByUserID(_ context.Context, u uuid.UUID) (*profile.ArtistProfile, error) {
	if p, ok := f.byUser[u]; ok {
		c := *p
		return &c, nil
	}
	return nil, apperrors.ErrNotFound
}

func (f *fakeProfiles) CreateDraftForUser(_ context.Context, u uuid.UUID, name string) (*profile.ArtistProfile, error) {
	p := &profile.ArtistProfile{ID: uuid.New(), UserID: u, DisplayName: name, Slug: "draft-" + name, Handle: "draft-" + name, Status: profile.ProfileStatusDraft}
	f.byUser[u] = p
	c := *p
	return &c, nil
}

func (f *fakeProfiles) SaveArtist(_ context.Context, p profile.ArtistProfile) (*profile.ArtistProfile, error) {
	f.saves++
	f.byUser[p.UserID] = &p
	c := p
	return &c, nil
}

type fakePosts struct {
	outbound.ArtPostRepository
	byID    map[uuid.UUID]art.ArtPost
	order   []uuid.UUID
	failOn  string
	creates int
}

func (f *fakePosts) GetByID(_ context.Context, id uuid.UUID) (*art.ArtPost, error) {
	if p, ok := f.byID[id]; ok {
		return &p, nil
	}
	return nil, apperrors.ErrNotFound
}

func (f *fakePosts) Create(_ context.Context, p art.ArtPost) (*art.ArtPost, error) {
	if p.Title == f.failOn {
		return nil, errors.New("db hiccup")
	}
	f.creates++
	f.byID[p.ID] = p
	f.order = append(f.order, p.ID)
	return &p, nil
}

func newProvisioner() (*ArtistProvisioner, *fakeProfiles, *fakePosts) {
	profiles := &fakeProfiles{byUser: map[uuid.UUID]*profile.ArtistProfile{}}
	posts := &fakePosts{byID: map[uuid.UUID]art.ArtPost{}}
	p := NewArtistProvisioner(profiles, posts)
	p.now = func() time.Time { return time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC) }
	return p, profiles, posts
}

func application(works ...string) onboarding.OnboardingApplication {
	app := onboarding.OnboardingApplication{ID: uuid.New(), ApplicantID: uuid.New(), DisplayName: "Ada", RequestedHandle: "ada_1"}
	year := 2024
	for i, title := range works {
		app.Works = append(app.Works, onboarding.ApplicantWork{
			ID: uuid.New(), Title: title, Medium: "oil", Year: &year, Position: i,
			URL: "https://res.cloudinary.com/c/image/upload/v1/mq/" + title + ".webp",
		})
	}
	return app
}

func TestApprovalMakesTheProfileLiveWithTheRequestedHandle(t *testing.T) {
	p, profiles, _ := newProvisioner()
	app := application("One", "Two")
	if err := p.ProvisionApproved(context.Background(), app); err != nil {
		t.Fatal(err)
	}
	got := profiles.byUser[app.ApplicantID]
	if got.Status != profile.ProfileStatusApproved || !got.IsPublic() {
		t.Fatalf("an approved artist must be public, status = %s", got.Status)
	}
	if got.Handle != "ada_1" || got.Slug != "ada_1" || got.ApprovedAt == nil {
		t.Fatalf("handle/slug/approved_at not set: %+v", got)
	}
}

func TestApprovalPublishesAnArtworkForEachSubmittedWork(t *testing.T) {
	p, profiles, posts := newProvisioner()
	app := application("One", "Two", "Three", "Four", "Five")
	if err := p.ProvisionApproved(context.Background(), app); err != nil {
		t.Fatal(err)
	}
	if len(posts.byID) != 5 {
		t.Fatalf("want 5 artworks, got %d", len(posts.byID))
	}
	owner := profiles.byUser[app.ApplicantID].ID
	first := posts.byID[WorkPostID(app.Works[0].ID)]
	if first.ArtistID != owner || first.Status != art.ArtStatusPublished || first.PublishedAt == nil {
		t.Fatalf("artwork not published for the artist: %+v", first)
	}
	if first.Title != "One" || first.Medium != "oil" || *first.Year != 2024 {
		t.Fatalf("details were not carried over: %+v", first)
	}
	if len(first.Media) != 1 || first.Media[0].URL != app.Works[0].URL {
		t.Fatalf("the uploaded image should be the artwork's media: %+v", first.Media)
	}
	// The first work is the newest, so it leads the page.
	if !first.PublishedAt.After(*posts.byID[WorkPostID(app.Works[4].ID)].PublishedAt) {
		t.Error("the first work should be published most recently")
	}
}

func TestProvisioningTwiceDoesNotDuplicate(t *testing.T) {
	p, profiles, posts := newProvisioner()
	app := application("One", "Two", "Three")
	for i := 0; i < 2; i++ {
		if err := p.ProvisionApproved(context.Background(), app); err != nil {
			t.Fatal(err)
		}
	}
	if len(posts.byID) != 3 || posts.creates != 3 {
		t.Fatalf("a repeat must not duplicate artworks: %d stored, %d creates", len(posts.byID), posts.creates)
	}
	if profiles.saves != 1 {
		t.Errorf("an unchanged profile should not be saved again, saves = %d", profiles.saves)
	}
}

func TestARetryAfterAPartialFailureFinishesTheRest(t *testing.T) {
	p, _, posts := newProvisioner()
	app := application("One", "Two", "Three")
	posts.failOn = "Two"
	if err := p.ProvisionApproved(context.Background(), app); err == nil {
		t.Fatal("the failure must be reported so the approval is rolled back to pending")
	}
	if len(posts.byID) != 1 {
		t.Fatalf("only the first artwork should exist so far, got %d", len(posts.byID))
	}
	posts.failOn = ""
	if err := p.ProvisionApproved(context.Background(), app); err != nil {
		t.Fatal(err)
	}
	if len(posts.byID) != 3 {
		t.Fatalf("the retry should finish all three, got %d", len(posts.byID))
	}
}

func TestAnExistingDraftProfileIsPublishedNotReplaced(t *testing.T) {
	p, profiles, _ := newProvisioner()
	app := application("One")
	existing := &profile.ArtistProfile{ID: uuid.New(), UserID: app.ApplicantID, Slug: "old", Handle: "old", DisplayName: "Ada", Bio: "Keep me", Status: profile.ProfileStatusDraft}
	profiles.byUser[app.ApplicantID] = existing
	if err := p.ProvisionApproved(context.Background(), app); err != nil {
		t.Fatal(err)
	}
	got := profiles.byUser[app.ApplicantID]
	if got.ID != existing.ID || got.Bio != "Keep me" || got.Status != profile.ProfileStatusApproved || got.Handle != "ada_1" {
		t.Fatalf("expected the same profile, published, with the new handle: %+v", got)
	}
}

func TestAnApplicationWithoutWorksStillGoesLive(t *testing.T) {
	p, profiles, posts := newProvisioner()
	app := application() // made before works were required
	if err := p.ProvisionApproved(context.Background(), app); err != nil {
		t.Fatal(err)
	}
	if !profiles.byUser[app.ApplicantID].IsPublic() || len(posts.byID) != 0 {
		t.Fatal("older applications go live with no artworks")
	}
}

func TestWorkPostIDIsStable(t *testing.T) {
	id := uuid.New()
	if WorkPostID(id) != WorkPostID(id) || WorkPostID(id) == WorkPostID(uuid.New()) {
		t.Fatal("the artwork id must depend only on the work id")
	}
}
