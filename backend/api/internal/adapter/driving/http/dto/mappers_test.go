package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

func TestToOnboardingApplicationResponse(t *testing.T) {
	now := time.Now().UTC()
	reviewer := uuid.New()
	app := onboarding.OnboardingApplication{
		ID:            uuid.New(),
		ApplicantID:   uuid.New(),
		ApplicantType: onboarding.ApplicantTypeArtist,
		DisplayName:   "Studio X",
		Notes:         "notes",
		Status:        onboarding.ApprovalStatusPending,
		ReviewedBy:    &reviewer,
		ReviewedAt:    &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	got := ToOnboardingApplicationResponse(app)
	assist.Equal(t, "artist", got.ApplicantType)
	assist.Equal(t, "pending", got.Status)
	assist.Equal(t, reviewer, *got.ReviewedBy)
}

func TestToOnboardingApplicationResponses(t *testing.T) {
	got := ToOnboardingApplicationResponses([]onboarding.OnboardingApplication{
		{DisplayName: "A", Status: onboarding.ApprovalStatusPending},
	})
	assist.Len(t, 1, len(got))
	assist.Equal(t, "A", got[0].DisplayName)
}

func TestToEventResponse_withLocationAndScrapedAt(t *testing.T) {
	now := time.Now().UTC()
	reviewer := uuid.New()
	image := "https://example.com/img.jpg"
	locID := uuid.New()
	event := events.Event{
		ID:          uuid.New(),
		Slug:        "opening",
		Title:       "Opening",
		Description: "desc",
		SourceURL:   "https://example.com",
		ImageURL:    &image,
		EventType:   "exhibition",
		Venue:       "Gallery",
		City:        "Addis",
		Location:    &events.Location{ID: locID, Name: "Main Hall", PinCoords: []float64{9.0, 38.7}},
		StartsAt:    now,
		Status:      events.EventStatusPending,
		ReviewNotes: "check",
		ReviewedBy:  &reviewer,
		ReviewedAt:  &now,
		ScrapedAt:   now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	got := ToEventResponse(event)
	assist.Equal(t, "pending", got.Status)
	assist.NotNil(t, got.Location)
	assist.Equal(t, "Main Hall", got.Location.Name)
	assist.NotNil(t, got.ScrapedAt)
}

func TestToEventResponses(t *testing.T) {
	got := ToEventResponses([]events.Event{{Title: "Show", StartsAt: time.Now().UTC()}})
	assist.Len(t, 1, len(got))
}

func TestToInstitutionResponse(t *testing.T) {
	now := time.Now().UTC()
	inst := institution.Institution{
		ID:          uuid.New(),
		Slug:        "gallery",
		Name:        "Gallery",
		Description: "A gallery",
		Contact: institution.ContactInfo{
			Email: "info@gallery.com",
		},
		Status:    institution.InstitutionStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}

	got := ToInstitutionResponse(inst)
	assist.Equal(t, "approved", got.Status)
	assist.Equal(t, "info@gallery.com", got.Contact.Email)
}

func TestToArtistProfileResponse(t *testing.T) {
	now := time.Now().UTC()
	p := profile.ArtistProfile{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Slug:        "artist",
		Handle:      "artist",
		DisplayName: "Artist",
		Featured:    true,
		Contact:     profile.ContactInfo{Email: "a@example.com"},
		Social:      profile.SocialLinks{Instagram: "@artist"},
		Status:      profile.ProfileStatusApproved,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	got := ToArtistProfileResponse(p)
	assist.Equal(t, "approved", got.Status)
	assist.Equal(t, "@artist", got.Social.Instagram)
}

func TestToArtistProfileResponses(t *testing.T) {
	got := ToArtistProfileResponses([]profile.ArtistProfile{{DisplayName: "A"}})
	assist.Len(t, 1, len(got))
}

func TestUpdateArtistProfileRequest_ToOwnProfileUpdate(t *testing.T) {
	req := UpdateArtistProfileRequest{
		DisplayName: "New Name",
		Slug:        "new-name",
		Handle:      "new-name",
		Bio:         "bio",
		Status:      "pending",
		Contact:     ContactInfoRequest{Email: "a@example.com"},
		Social:      SocialLinksRequest{Twitter: "@x"},
	}

	got := req.ToOwnProfileUpdate()
	assist.Equal(t, "New Name", got.DisplayName)
	assist.Equal(t, profile.ProfileStatusPending, got.Status)
}

func TestArtPostMappers(t *testing.T) {
	now := time.Now().UTC()
	year := 2024
	post := art.ArtPost{
		ID:       uuid.New(),
		ArtistID: uuid.New(),
		Title:    "Work",
		Medium:   "oil",
		Year:     &year,
		Media: []art.MediaAsset{{
			ID: uuid.New(), URL: "https://example.com/a.jpg", SortOrder: 1,
		}},
		Status:    art.ArtStatusPublished,
		CreatedAt: now,
		UpdatedAt: now,
	}

	single := ToArtPostResponse(post)
	assist.Equal(t, "published", single.Status)
	assist.Len(t, 1, len(single.Media))

	many := ToArtPostResponses([]art.ArtPost{post})
	assist.Len(t, 1, len(many))

	withArtist := ToArtPostWithArtistResponse(art.ArtPostWithArtist{
		ArtPost:    post,
		ArtistSlug: "artist",
		ArtistName: "Artist",
	})
	assist.Equal(t, "artist", withArtist.ArtistSlug)

	withArtists := ToArtPostWithArtistResponses([]art.ArtPostWithArtist{{ArtPost: post, ArtistSlug: "artist"}})
	assist.Len(t, 1, len(withArtists))
}
